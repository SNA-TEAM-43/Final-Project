package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	maxLoadWorkers      = 32
	maxLoadQueryCount   = 100_000
	maxLoadDurationMS   = 300_000 // 5 minutes
	maxLoadPayloadSize  = 1 << 20 // 1 MiB
	maxLoadRowsPerQuery = 1_000
)

type loadCPURequest struct {
	DurationMS int     `json:"duration_ms"`
	Workers    int     `json:"workers"`
	Intensity  float64 `json:"intensity"`
}

type loadCPUResponse struct {
	DurationMS int     `json:"duration_ms"`
	Workers    int     `json:"workers"`
	Intensity  float64 `json:"intensity"`
	ElapsedMS  int64   `json:"elapsed_ms"`
}

type dbWriteParams struct {
	QueryCount  int `json:"query_count"`
	Workers     int `json:"workers"`
	PayloadSize int `json:"payload_size"`
}

type dbReadParams struct {
	QueryCount   int    `json:"query_count"`
	Workers      int    `json:"workers"`
	Mode         string `json:"mode"`
	RowsPerQuery int    `json:"rows_per_query"`
}

type dbLoadResult struct {
	ElapsedMS     int64  `json:"elapsed_ms"`
	QueriesOK     int64  `json:"queries_ok"`
	QueriesFailed int64  `json:"queries_failed"`
	LastError     string `json:"last_error"`
}

type loadCleanupResult struct {
	Deleted int64 `json:"deleted"`
}

func printCPUResult(res loadCPUResponse) {
	fmt.Printf("	elapsed: %d ms\n", res.ElapsedMS)
	fmt.Printf("	workers: %d\n", res.Workers)
	fmt.Printf("	intensity: %f\n", res.Intensity)
}

func printDBResult(res dbLoadResult) {
	fmt.Printf("	elapsed: %d ms\n", res.ElapsedMS)
	fmt.Printf("	ok: %d  failed: %d\n", res.QueriesOK, res.QueriesFailed)
	if res.LastError != "" {
		fmt.Printf("	last_error: %s\n", res.LastError)
	}
}

func mediumCPUParams() loadCPURequest {
	return loadCPURequest{DurationMS: 3000, Workers: 4, Intensity: 0.9}
}

func killCPUParams() loadCPURequest {
	return loadCPURequest{DurationMS: maxLoadDurationMS, Workers: maxLoadWorkers, Intensity: 1.0}
}

func mediumDBWriteParams() dbWriteParams {
	return dbWriteParams{QueryCount: 500, Workers: 8, PayloadSize: 4096}
}

func killDBWriteParams() dbWriteParams {
	return dbWriteParams{QueryCount: maxLoadQueryCount, Workers: maxLoadWorkers, PayloadSize: maxLoadPayloadSize}
}

func mediumDBReadParams() dbReadParams {
	return dbReadParams{QueryCount: 200, Workers: 4, Mode: "count", RowsPerQuery: 10}
}

func killDBReadParams(mode string) dbReadParams {
	return dbReadParams{QueryCount: maxLoadQueryCount, Workers: maxLoadWorkers, Mode: mode, RowsPerQuery: maxLoadRowsPerQuery}
}

func runCPU(c *Client) error {
	return runCPUWith(c, loadCPURequest{DurationMS: 1000, Workers: 2, Intensity: 1})
}

func runCPUWith(c *Client, req loadCPURequest) error {
	var res loadCPUResponse
	if err := c.PostJSON("/load/cpu", req, &res); err != nil {
		return err
	}
	printCPUResult(res)
	return nil
}

func runDBWrite(c *Client) error {
	return runDBWriteWith(c, dbWriteParams{QueryCount: 100, Workers: 4, PayloadSize: 256})
}

func runDBWriteWith(c *Client, req dbWriteParams) error {
	var res dbLoadResult
	if err := c.PostJSON("/load/db/write", req, &res); err != nil {
		return err
	}
	printDBResult(res)
	return nil
}

func runDBRead(c *Client) error {
	return runDBReadWith(c, dbReadParams{QueryCount: 50, Workers: 2, Mode: "latest", RowsPerQuery: 10})
}

func runDBReadWith(c *Client, req dbReadParams) error {
	var res dbLoadResult
	if err := c.PostJSON("/load/db/read", req, &res); err != nil {
		return err
	}
	printDBResult(res)
	return nil
}

func runCleanup(c *Client) error {
	var res loadCleanupResult
	if err := c.DeleteJSON("/load/db/jobs", &res); err != nil {
		return err
	}
	fmt.Printf("	deleted: %d\n", res.Deleted)
	return nil
}

func runStress(c *Client) error {
	fmt.Println("STRESS: cpu + db write + db read (parallel)")
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	runOne := func(name string, fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("  [%s] start\n", name)
			if err := fn(); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", name, err))
				mu.Unlock()
				fmt.Printf("  [%s] fail: %v\n", name, err)
				return
			}
			fmt.Printf("  [%s] done\n", name)
		}()
	}

	runOne("cpu", func() error { return runCPUWith(c, mediumCPUParams()) })
	runOne("db_write", func() error { return runDBWriteWith(c, mediumDBWriteParams()) })
	runOne("db_read", func() error { return runDBReadWith(c, mediumDBReadParams()) })

	wg.Wait()
	return errors.Join(errs...)
}

type chaosStatus struct {
	Active      bool  `json:"active"`
	RemainingMS int64 `json:"remaining_ms"`
}

func runChaos(c *Client) error {
	fmt.Println("CHAOS: injecting DB write failures for 30 seconds")
	var res map[string]any
	if err := c.PostJSON("/chaos/db-fail?duration_ms=30000", nil, &res); err != nil {
		return err
	}
	fmt.Println("  chaos active — firing DB writes in background for 30s (all will fail)")
	fmt.Println("  watch: db_queries_total{status=\"failed\"} in Grafana")

	go func() {
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			runDBWriteWith(c, dbWriteParams{QueryCount: 50, Workers: 4, PayloadSize: 256})
		}
		fmt.Println("  [chaos] background load done")
	}()

	return nil
}

func runChaosHeal(c *Client) error {
	var res map[string]any
	if err := c.PostJSON("/chaos/heal", nil, &res); err != nil {
		return err
	}
	fmt.Println("  chaos cancelled — DB writes restored")
	return nil
}

func runKill(c *Client, confirm string) error {
	if confirm != "y" && confirm != "Y" {
		fmt.Println("cancelled")
		return nil
	}

	fmt.Println("KILL: max load on cpu + db write + db read (parallel)")
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	runOne := func(name string, fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("  [%s] start\n", name)
			if err := fn(); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", name, err))
				mu.Unlock()
				fmt.Printf("  [%s] fail: %v\n", name, err)
				return
			}
			fmt.Printf("  [%s] done\n", name)
		}()
	}

	runOne("cpu", func() error { return runCPUWith(c, killCPUParams()) })
	runOne("db_write", func() error { return runDBWriteWith(c, killDBWriteParams()) })
	runOne("db_read_count", func() error { return runDBReadWith(c, killDBReadParams("count")) })
	runOne("db_read_latest", func() error { return runDBReadWith(c, killDBReadParams("latest")) })

	wg.Wait()
	return errors.Join(errs...)
}

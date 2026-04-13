package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"performance-testing/internal/calculator"
	"time"
)

func main() {
	// 启动pprof服务器
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// 运行一些计算密集型任务
	log.Println("开始性能测试...")
	for i := 0; i < 10; i++ {
		start := time.Now()
		result := calculator.Fibonacci(30)
		duration := time.Since(start)
		log.Printf("Fibonacci(30) = %d, 耗时: %v\n", result, duration)
		time.Sleep(500 * time.Millisecond)
	}

	log.Println("性能测试完成，pprof服务器仍在运行，请访问 http://localhost:6060/debug/pprof/")
	select {}
}
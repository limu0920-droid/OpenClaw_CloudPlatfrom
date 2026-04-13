package main

import (
	"fmt"
	"log"
	"net/http"
	"performance-testing/internal/calculator"
	"strconv"
)

func fibHandler(w http.ResponseWriter, r *http.Request) {
	// 从查询参数获取n
	nStr := r.URL.Query().Get("n")
	n, err := strconv.Atoi(nStr)
	if err != nil || n < 0 {
		n = 20 // 默认值
	}

	// 计算斐波那契数
	result := calculator.Fibonacci(n)

	// 返回结果
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"n": %d, "result": %d}`, n, result)
}

func main() {
	http.HandleFunc("/fib", fibHandler)
	log.Println("服务器启动在 :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
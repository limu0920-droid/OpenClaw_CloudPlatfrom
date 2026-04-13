package calculator

import "testing"

// BenchmarkFibonacci 测试Fibonacci函数的性能
func BenchmarkFibonacci(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Fibonacci(20)
	}
}

// BenchmarkSum 测试Sum函数的性能
func BenchmarkSum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Sum(10000)
	}
}

// BenchmarkFactorial 测试Factorial函数的性能
func BenchmarkFactorial(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Factorial(10)
	}
}
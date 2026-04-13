package calculator

// Fibonacci 计算斐波那契数列（使用记忆化）
func Fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	memo := make(map[int]int)
	memo[0] = 0
	memo[1] = 1
	return fibonacciMemo(n, memo)
}

// fibonacciMemo 带记忆化的斐波那契计算
func fibonacciMemo(n int, memo map[int]int) int {
	if val, ok := memo[n]; ok {
		return val
	}
	memo[n] = fibonacciMemo(n-1, memo) + fibonacciMemo(n-2, memo)
	return memo[n]
}

// Sum 计算1到n的和（使用数学公式）
func Sum(n int) int {
	return n * (n + 1) / 2
}

// Factorial 计算阶乘（使用迭代）
func Factorial(n int) int {
	if n <= 1 {
		return 1
	}
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}
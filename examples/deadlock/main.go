// Command deadlock — настоящее круговое ожидание: goroutine-a ждёt то, что
// может отправить только goroutine-b, и наоборот. Ни один Send так и не
// случается, и рантайм Go обнаруживает тупик сам — процесс завершается с
// "fatal error: all goroutines are asleep - deadlock!". Это ожидаемо: сервер
// запускает пример как отдельный подпроцесс именно ради такой изоляции.
package main

import "github.com/LunatikDG/goscope/internal/instrument"

func main() {
	waitForA := make(chan struct{}) // здесь ждёт goroutine-b то, что мог бы прислать goroutine-a
	waitForB := make(chan struct{}) // здесь ждёт goroutine-a то, что мог бы прислать goroutine-b

	go goroutineA(waitForB)
	go goroutineB(waitForA)

	select {} // тоже блокируемся навсегда — вместе с a и b это полноценный тупик всей программы
}

func goroutineA(waitForB <-chan struct{}) {
	instrument.Spawn(1, "goroutine-a")
	instrument.Block(1, 2)
	<-waitForB
}

func goroutineB(waitForA <-chan struct{}) {
	instrument.Spawn(2, "goroutine-b")
	instrument.Block(2, 1)
	<-waitForA
}

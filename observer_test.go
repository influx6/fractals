package fractals_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/influx6/faux/context"
	"github.com/influx6/fractals"
)

func TestObserverEnding(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	ob := fractals.NewObservable(fractals.NewBehaviour(func(name string) string {
		return "Mr." + name
	}, nil, nil), false)

	ob2 := fractals.NewObservable(fractals.NewBehaviour(func(name string) string {
		wg.Done()
		return name + "!"
	}, nil, nil), false)

	obEnd := ob.Subscribe(ob2)

	ob.Next(context.New(), "Thunder")
	ob2.DoneVal(true)
	obEnd.End()
	ob.Next(context.New(), "Walkte")
	wg.Wait()

	ob.DoneVal(true)
	ob.End()
	ob2.End()
}

func TestDebounceObserver(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	ob := fractals.NewObservable(fractals.NewBehaviour(func(name string) string {
		return "Mr." + name
	}, nil, nil), false)

	ob2 := fractals.DebounceWithObserver(ob, 10*time.Millisecond)

	ob2.Subscribe(fractals.NewObservable(fractals.NewBehaviour(func(name string) {
		fmt.Printf("Debounce: %s\n", name)
		wg.Done()
	}, nil, nil), false))

	// These items wont be seen.
	ob.Next(context.New(), "Thunder")
	ob.Next(context.New(), "Thunder2")
	ob.Next(context.New(), "Thunder3")
	ob.Next(context.New(), "Thunder4")

	<-time.After(11 * time.Millisecond)
	ob.Next(context.New(), "Lightening")

	// These items wont be seen.
	ob.Next(context.New(), "Thunder")
	ob.Next(context.New(), "Thunder2")
	ob.Next(context.New(), "Thunder3")
	ob.Next(context.New(), "Thunder4")

	<-time.After(11 * time.Millisecond)
	ob.Next(context.New(), "Slickering")

	wg.Wait()
	ob.DoneVal(true)
	ob2.DoneVal(true)

	ob.End()
	ob2.End()
}

func TestObserver(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	ob := fractals.NewObservable(fractals.NewBehaviour(func(name string) string {
		return "Mr." + name
	}, nil, nil), false)

	ob2 := fractals.NewObservable(fractals.NewBehaviour(func(name string) string {
		wg.Done()
		return name + "!"
	}, nil, nil), false)

	ob2.Subscribe(fractals.NewObservable(fractals.NewBehaviour(func(name string) {
		wg.Done()
	}, nil, nil), false))

	ob.Subscribe(ob2)

	ob.Next(context.New(), "Thunder")
	wg.Wait()

	ob.DoneVal(true)
	ob2.DoneVal(true)

	ob.End()
	ob2.End()
}

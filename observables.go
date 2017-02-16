package fractals

import (
	"time"

	"github.com/influx6/faux/context"
	"github.com/influx6/faux/raf"
)

var identity = IdentityHandler()

// Behaviour exposes a struct which defines a type which allows creating a set of
// pure structure for calling behaviours based on reactions like a stream.
type Behaviour struct {
	Next Handler
	Done Handler
	End  func()
}

// NewBehaviour returns a new instance of a Behaviour struct.
func NewBehaviour(next, done interface{}, end func()) Behaviour {
	var b Behaviour
	b.Next = MustWrap(next)
	b.End = end

	if done != nil {
		b.Done = MustWrap(done)
	}

	return b
}

// IdentityBehaviour returns a Behaviour struct which exposes the behaviours which
// an observer is to perform.
func IdentityBehaviour() Behaviour {
	return Behaviour{
		Next: identity,
		Done: identity,
	}
}

// Observable defines a interface that provides a type by which continouse
// events stream can occur.
type Observable interface {
	End()
	Async() Observable
	Sync() Observable
	NextVal(interface{})
	DoneVal(interface{})
	Done(context.Context, interface{})
	Next(context.Context, interface{})
	AddFinalizer(func())
	Subscribe(Observable, ...func()) *Subscription
}

// NewObservable returns a new instance of a Observable.
func NewObservable(behaviour Behaviour, async bool) Observable {
	if behaviour.Next == nil {
		panic("No next Handler provided")
	}

	if behaviour.Done == nil {
		behaviour.Done = identity
	}

	return &IndefiniteObserver{
		behaviour: behaviour,
		doAsync:   async,
	}
}

// ReplayObservable returns a new instance of a Observable which replaces it's
// events down it's subscribers line.
func ReplayObservable() Observable {
	return &IndefiniteObserver{
		behaviour: IdentityBehaviour(),
	}
}

// MapWithObserver applies the giving predicate to all values the target observer
// provides returning only values which match.
func MapWithObserver(mapPredicate Behaviour, target Observable) Observable {
	ob := NewObservable(mapPredicate, false)
	target.Subscribe(ob, ob.End)
	return ob
}

// DebounceRAFWithObserver applies the giving predicate to all values the target observer
// provides returning only values which match. It attempts to use the RAF implementation
// to work.
func DebounceRAFWithObserver(target Observable) Observable {
	var allowed bool

	id := raf.RequestAnimationFrame(func(dt float64) {
		allowed = true
	})

	ob := NewObservable(Behaviour{
		Next: MustWrap(func(item interface{}) interface{} {
			if !allowed {
				return nil
			}

			allowed = false
			return item
		}),
	}, false)

	ob.AddFinalizer(func() {
		raf.CancelAnimationFrame(id)
	})

	target.Subscribe(ob)

	return ob
}

// DebounceWithObserver applies the giving predicate to all values the target observer
// provides returning only values which matches and uses the time.Ticker.
func DebounceWithObserver(target Observable, dr time.Duration) Observable {
	var allowed bool

	ticker := time.NewTicker(dr)

	go func() {
		for {
			_, open := <-ticker.C
			if !open {
				break
			}

			allowed = true
		}
	}()

	ob := NewObservable(Behaviour{
		Next: MustWrap(func(item interface{}) interface{} {
			if !allowed {
				return nil
			}

			allowed = false
			return item
		}),
	}, false)

	ob.AddFinalizer(func() {
		if ticker != nil {
			ticker.Stop()
			ticker = nil
		}
	})

	target.Subscribe(ob)

	return ob
}

// FilterWithObserver applies the giving predicate to all values the target observer
// provides returning only values which match.
func FilterWithObserver(predicate func(interface{}) bool, target Observable) Observable {
	ob := NewObservable(Behaviour{
		Next: MustWrap(func(item interface{}) interface{} {
			if predicate(item) {
				return item
			}

			return nil
		}),
	}, false)

	target.Subscribe(ob)
	return ob
}

// IndefiniteObserver defines a structure which implements the concrete structure
// of the Observable interface. It provides a baseline interface which others
// can inherit from.
type IndefiniteObserver struct {
	behaviour  Behaviour
	subs       []*Subscription
	finalizers []func()
	doAsync    bool
}

// Subscribe connects the giving Observer with the provide observer and returns a
// subscription object which disconnects the giving event stream.
func (in *IndefiniteObserver) Subscribe(b Observable, finalizers ...func()) *Subscription {
	var sub Subscription
	sub.observer = b
	sub.handlers = finalizers

	in.subs = append(in.subs, &sub)

	return &sub
}

// Subscription defines the structure which holds the connection between two
// observers.
type Subscription struct {
	observer Observable
	handlers []func()
}

// End defines a function to disconnect the observer from a giving subscription.
func (sub *Subscription) End() {
	sub.observer = nil

	// Run finalizers for subscription.
	for _, fl := range sub.handlers {
		fl()
	}
}

// End discloses all subscription to the observer, calling their appropriate
// finalizers.
func (in *IndefiniteObserver) End() {
	defer func() {
		in.finalizers = nil
		in.subs = nil
	}()

	for _, fl := range in.finalizers {
		fl()
	}

	for _, sub := range in.subs {
		if sub.observer == nil {
			continue
		}

		sub.End()
	}
}

// AddFinalizer adds a giving finalizer which will be runned when the giving
// observer has ended.
func (in *IndefiniteObserver) AddFinalizer(val func()) {
	in.finalizers = append(in.finalizers, val)
}

// NextVal receives the value to be passed to the Observer.Next function and
// creates a new context for call.
func (in *IndefiniteObserver) NextVal(val interface{}) {
	in.Next(context.New(), val)
}

// Next receives the next input for the observer to run it's internal
// calls against and which then passes to all it's next subscribers.
func (in *IndefiniteObserver) Next(ctx context.Context, val interface{}) {
	if in.doAsync {
		go func() {
			var err error
			var res interface{}

			if errx, ok := val.(error); ok {
				res, err = in.behaviour.Next(ctx, errx, nil)
			} else {
				res, err = in.behaviour.Next(ctx, nil, val)
			}

			for _, sub := range in.subs {
				if sub.observer == nil {
					continue
				}

				if err != nil {
					sub.observer.Next(ctx, err)
					continue
				}

				sub.observer.Next(ctx, res)
			}
		}()
		return
	}

	var err error
	var res interface{}

	if errx, ok := val.(error); ok {
		res, err = in.behaviour.Next(ctx, errx, nil)
	} else {
		res, err = in.behaviour.Next(ctx, nil, val)
	}

	for _, sub := range in.subs {
		if sub.observer == nil {
			continue
		}

		if err != nil {
			sub.observer.Next(ctx, err)
			continue
		}

		sub.observer.Next(ctx, res)
	}
}

// DoneVal receives the value to be passed to the Observer.Done function and
// creates a new context for call.
func (in *IndefiniteObserver) DoneVal(val interface{}) {
	in.Done(context.New(), val)
}

// Done receives the done input for the observer to run it's internal
// calls against and which then passes to all it's next subscribers.
func (in *IndefiniteObserver) Done(ctx context.Context, val interface{}) {
	if in.doAsync {
		go func() {
			var err error
			var res interface{}

			if errx, ok := val.(error); ok {
				res, err = in.behaviour.Done(ctx, errx, nil)
			} else {
				res, err = in.behaviour.Done(ctx, nil, val)
			}

			for _, sub := range in.subs {
				if sub.observer == nil {
					continue
				}

				if err != nil {
					sub.observer.Done(ctx, err)
					continue
				}

				sub.observer.Done(ctx, res)
			}

		}()
		return
	}

	var err error
	var res interface{}

	if errx, ok := val.(error); ok {
		res, err = in.behaviour.Done(ctx, errx, nil)
	} else {
		res, err = in.behaviour.Done(ctx, nil, val)
	}

	for _, sub := range in.subs {
		if sub.observer == nil {
			continue
		}

		if err != nil {
			sub.observer.Done(ctx, err)
			continue
		}

		sub.observer.Done(ctx, res)
	}
}

// Async returns a new observer which runs its behaviour in a goroutine to provide
// asynchronouse processing. No copy is made, all subscriptions are left intact
// with the core synchronouse version. Any effect which occurs with this version
// occurs with the asynchronouse version. This is an intended effect.
func (in *IndefiniteObserver) Async() Observable {
	if in.doAsync {
		return in
	}

	return &IndefiniteObserver{
		behaviour: in.behaviour,
		subs:      in.subs[:len(in.subs)],
		doAsync:   true,
	}
}

// Sync returns a new observer which runs its behaviour in a goroutine to provide
// non-asynchronouse processing. No copy is made, all subscriptions are left intact
// with the first asynchronouse version. Any effect which occurs with this version
// occurs with the non-asynchronouse version. This is an intended effect.
func (in *IndefiniteObserver) Sync() Observable {
	if !in.doAsync {
		return in
	}

	return &IndefiniteObserver{
		behaviour: in.behaviour,
		subs:      in.subs[:len(in.subs)],
		doAsync:   false,
	}
}

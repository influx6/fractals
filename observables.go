package fractals

import "github.com/influx6/faux/context"

// Observable defines a interface that provides a type by which continouse
// events stream can occur.
type Observable interface {
	End()
	Next(context.Context, interface{})
	Subscribe(Observable, ...func()) Subscription
}

// NewObservable returns a new instance of a Observable.
func NewObservable(behaviour interface{}) Observable {
	return &IndefiniteObserver{
		onNext: MustWrap(behaviour),
	}
}

// Subscription defines a structure which provides a subscription handle for which
// an observer recieves when registered on a subscription.
type Subscription interface {
	End()
}

type subscription struct {
	observer   Observable
	finalizers []func()
}

// Finalize ends and runs all ending mechanisms required before ending the
// subscriptions
func (sub *subscription) Finalize() {
	for _, fn := range sub.finalizers {
		fn()
	}
}

// End defines a function to disconnect the observer from a giving subscription.
func (sub *subscription) End() {
	sub.observer = nil
	sub.Finalize()
}

// IndefiniteObserver defines a structure which implements the concrete structure
// of the Observable interface. It provides a baseline interface which others
// can inherit from.
type IndefiniteObserver struct {
	onNext Handler
	subs   []*subscription
}

// Subscribe connects the giving Observer with the provide observer and returns a
// subscription object which disconnects the giving event stream.
func (in *IndefiniteObserver) Subscribe(b Observable, finalizers ...func()) Subscription {
	var sub subscription
	sub.observer = b
	sub.finalizers = finalizers

	in.subs = append(in.subs, &sub)

	return &sub
}

// End discloses all subscription to the observer, calling their appropriate
// finalizers.
func (in *IndefiniteObserver) End() {
	for _, sub := range in.subs {
		if sub.observer == nil {
			continue
		}

		sub.End()
	}
}

// Next receives the next input for the observer to run it's internal
// calls against and which then passes to all it's next subscribers.
func (in *IndefiniteObserver) Next(ctx context.Context, val interface{}) {
	var err error
	var res interface{}

	if errx, ok := val.(error); ok {
		res, err = in.onNext(ctx, errx, nil)
	} else {
		res, err = in.onNext(ctx, nil, val)
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

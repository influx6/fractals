package fractals

import "github.com/influx6/faux/context"

// Observable defines a interface that provides a type by which continouse
// events stream can occur.
type Observable interface {
	End()
	AddFinalizers(...func())
	Next(context.Context, interface{})
	Subscribe(Observable, ...func()) Subscription
}

// NewObservable returns a new instance of a Observable.
func NewObservable(behaviour interface{}, finalizers ...func()) Observable {
	return &IndefiniteObserver{
		onNext:     MustWrap(behaviour),
		finalizers: finalizers,
	}
}

// ReplayObservable returns a new instance of a Observable which replaces it's
// events down it's subscribers line.
func ReplayObservable(finalizers ...func()) Observable {
	return &IndefiniteObserver{
		onNext:     IdentityHandler(),
		finalizers: finalizers,
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
	onNext     Handler
	subs       []*subscription
	finalizers []func() //pure functions which should perform some cleanup.
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

	in.finalize()
}

// AddFinalizers adds the the sets of pure functions to be called once the
// observers End(), function is called. This allows clean up operations to be
// performed if required.
func (in *IndefiniteObserver) AddFinalizers(fx ...func()) {
	in.finalizers = append(in.finalizers, fx...)
}

// finalize ends and runs all ending functions to perform any cleanup for the
// observer.
func (in *IndefiniteObserver) finalize() {
	for _, fn := range in.finalizers {
		fn()
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

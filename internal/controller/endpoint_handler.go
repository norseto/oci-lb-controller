/*
MIT License

Copyright (c) 2024 Norihiro Seto

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package controller

import (
	"context"

	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

// EndpointHandler is a struct that implements the EventHandler interface.
type EndpointHandler struct {
	client.Client
	Recorder record.EventRecorder
}

func (eh *EndpointHandler) Create(ctx context.Context, evt event.CreateEvent, _ workqueue.RateLimitingInterface) {
}

func (eh *EndpointHandler) Update(ctx context.Context, evt event.UpdateEvent, _ workqueue.RateLimitingInterface) {
}

func (eh *EndpointHandler) Delete(ctx context.Context, evt event.DeleteEvent, _ workqueue.RateLimitingInterface) {
}

func (eh *EndpointHandler) Generic(context.Context, event.GenericEvent, workqueue.RateLimitingInterface) {
	// Do nothing
}

package tcp_service

type IRouter interface {
	PreHandler(IRequest)
	Handler(IRequest)
	PostHandler(IRequest)
}

type BaseRouter struct {
}

func (r *BaseRouter) PreHandler(request IRequest) {
}

func (r *BaseRouter) Handler(request IRequest) {
}

func (r *BaseRouter) PostHandler(request IRequest) {
}

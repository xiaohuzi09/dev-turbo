package service

// GreetService 提供问候服务
type GreetService struct{}

// Greet 返回问候语
func (g *GreetService) Greet(name string) string {
	return "Hello " + name + "!"
}

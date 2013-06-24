package main
import (
    "fmt"
    "time"
)

type SetterClosure func(res interface{})
type GetterClosure func(res interface{}) interface{}

type Actor struct {
    setterChan chan <- SetterClosure
    getterChan chan <- GetterClosure
    getterRetChan <-chan interface{}
}
func NewActor(res interface{}, buffSize int) *Actor {

    setterChan, getterChan := make(chan SetterClosure, buffSize), make(chan GetterClosure, buffSize)
    getterRetChan := make(chan interface{})

    go func() {
        l_for: for {
            select {
            case setter, ok := <- setterChan:
                if !ok { break l_for }
                setter(res)
            case getter := <- getterChan:
                getterRetChan <- getter(res)
            }
        }

        close(getterRetChan)
    }()

    return &Actor{setterChan, getterChan, getterRetChan}
}
func (this *Actor) DoSet(setter SetterClosure) {
    this.setterChan <- setter
}
func (this *Actor) DoGet(getter GetterClosure) interface{} {
    this.getterChan <- getter
    return <- this.getterRetChan
}
func (this *Actor) Destroy() {
    close(this.setterChan)
    <-this.getterRetChan
    close(this.getterChan)
}

func main() {

    actor := NewActor(new(int), 1)

    for i := 0; i < 10; i++ { 
        go func() {
            for i := 0; i < 10000; i++ {
                actor.DoSet(func(res interface{}){
                    *res.(*int)--
                })
            }
        }()
    }
    for i := 0; i < 10; i++ { 
        go func() {
            for i := 0; i < 10001; i++ {
                actor.DoSet(func(res interface{}){
                    *res.(*int)++
                })
            }
        }()
    }

    time.Sleep(time.Second * 1)

    n := actor.DoGet(func(res interface{}) interface{}{
        return *res.(*int)
    }).(int)

    actor.Destroy()
    fmt.Println(n)
}

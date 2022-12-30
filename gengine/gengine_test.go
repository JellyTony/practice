package gengine

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/bilibili/gengine/builder"
	"github.com/bilibili/gengine/context"
	"github.com/bilibili/gengine/engine"
	"github.com/stretchr/testify/assert"
)

type User struct {
	Name string
	Age  int64
	Male bool
}

func (u *User) SayHi(s string) {
	fmt.Println("Hi " + s + ", I am " + u.Name)
}

func PrintAge(age int64) {
	fmt.Println("Age is " + strconv.FormatInt(age, 10))
}

// TestSingleRule 单个规则
func TestSingleRule(t *testing.T) {
	const (
		rule1 = `
rule "rule1" "a test" salience 10
begin
   println(@name)
   user.SayHi("lily")
   if user.Age > 20{
      newAge = user.Age + 100
      user.Age = newAge
   }
   PrintAge(user.Age)
   user.Male = false
end `
	)

	dataContext := context.NewDataContext()
	user := &User{
		Name: "Calo",
		Age:  25,
		Male: true,
	}
	dataContext.Add("user", user)
	dataContext.Add("println", fmt.Println)
	dataContext.Add("PrintAge", PrintAge)
	ruleBuilder := builder.NewRuleBuilder(dataContext)
	err := ruleBuilder.BuildRuleFromString(rule1)
	assert.NoError(t, err)
	eng := engine.NewGengine()

	err = eng.Execute(ruleBuilder, true)
	assert.NoError(t, err)

	t.Logf("Age=%d Name=%s,Male=%t", user.Age, user.Name, user.Male)
}

type Station struct {
	Temperature  int64 // 温度
	Humidity     int64 // 湿度
	Water        int64 // 水浸
	Smoke        int64 // 烟雾
	Door1        int64 // 门禁1
	Door2        int64 // 门禁2
	StationState int64 // 探测站状态:   0正常；1预警；2异常；3未知
}

const (
	stateRule = `
rule "normalRule" "探测站状态正常计算规则" salience 8
begin
   println("/***************** 正常规则 ***************")
   if Station.Temperature>0 && Station.Temperature<80 && Station.Humidity<70 && Station.Water==0 && Station.Smoke==0 && Station.Door1==0 && Station.Door2==0{
      Station.StationState=0
      println("满足")
   }else{
      println("不满足")
   }
end
 
rule "errorRule" "探测站状态预警计算规则" salience 9
begin
   println("/***************** 预警规则 ***************")
   if Station.Temperature>0 && Station.Temperature<80 && Station.Humidity<70 && Station.Water==0 && Station.Smoke==0 && (Station.Door1==1 || Station.Door2==1){
      Station.StationState=1
      println("满足")
   }else{
      println("不满足")
   }
end
 
rule "warnRule" "探测站状态异常计算规则" salience 10
begin
   println("/***************** 异常规则 ***************")
   if Station.Temperature<0 || Station.Temperature>80 || Station.Humidity>70 || Station.Water==1 || Station.Smoke==1{
      Station.StationState=2
      println("满足")
   }else{
      println("不满足")
   }
end `
)

// TestSequentialExecution 顺序执行
func TestSequentialExecution(t *testing.T) {
	station := &Station{
		Temperature:  40,
		Humidity:     30,
		Water:        0,
		Smoke:        1,
		Door1:        0,
		Door2:        1,
		StationState: 0,
	}
	dataContext := context.NewDataContext()
	dataContext.Add("Station", station)
	dataContext.Add("println", fmt.Println)
	ruleBuilder := builder.NewRuleBuilder(dataContext)
	err := ruleBuilder.BuildRuleFromString(stateRule)
	assert.NoError(t, err)
	eng := engine.NewGengine()
	err = eng.Execute(ruleBuilder, true)
	assert.NoError(t, err)
	t.Logf("StationState=%d", station.StationState)
}

type Temperature struct {
	Tag   string  // 标签点名称
	Value float64 // 数据值
	State int64   // 状态
	Event string  // 报警事件
}

type Water struct {
	Tag   string // 标签点名称
	Value int64  // 数据值
	State int64  // 状态
	Event string // 报警事件
}

type Smoke struct {
	Tag   string // 标签点名称
	Value int64  // 数据值
	State int64  // 状态
	Event string // 报警事件
}

const (
	eventRule = `
rule "TemperatureRule" "温度事件计算规则"
begin
   println("/***************** 温度事件计算规则 ***************/")
   tempState = 0
   if Temperature.Value < 0{
      tempState = 1
   }else if Temperature.Value > 80{
      tempState = 2
   }
   if Temperature.State != tempState{
      if tempState == 0{
         Temperature.Event = "温度正常"
      }else if tempState == 1{
         Temperature.Event = "低温报警"
      }else{
         Temperature.Event = "高温报警"
      }
   }else{
      Temperature.Event = ""
   }
   Temperature.State = tempState
end
 
rule "WaterRule" "水浸事件计算规则"
begin
   println("/***************** 水浸事件计算规则 ***************/")
   tempState = 0
   if Water.Value != 0{
      tempState = 1
   }
   if Water.State != tempState{
      if tempState == 0{
         Water.Event = "水浸正常"
      }else{
         Water.Event = "水浸异常"
      }
   }else{
      Water.Event = ""
   }
   Water.State = tempState
end
 
rule "SmokeRule" "烟雾事件计算规则"
begin
   println("/***************** 烟雾事件计算规则 ***************/")
   tempState = 0
   if Smoke.Value != 0{
      tempState = 1
   }
   if Smoke.State != tempState{
      if tempState == 0{
         Smoke.Event = "烟雾正常"
      }else{
         Smoke.Event = "烟雾报警"
      }
   }else{
      Smoke.Event = ""
   }
   Smoke.State = tempState
end
`
)

// TestConcurrentExecution 并发执行
func TestConcurrentExecution(t *testing.T) {
	temperature := &Temperature{
		Tag:   "temperature",
		Value: 90,
		State: 0,
		Event: "",
	}
	water := &Water{
		Tag:   "water",
		Value: 0,
		State: 0,
		Event: "",
	}
	smoke := &Smoke{
		Tag:   "smoke",
		Value: 1,
		State: 0,
		Event: "",
	}
	dataContext := context.NewDataContext()
	dataContext.Add("Temperature", temperature)
	dataContext.Add("Water", water)
	dataContext.Add("Smoke", smoke)
	dataContext.Add("println", fmt.Println)
	ruleBuilder := builder.NewRuleBuilder(dataContext)
	err := ruleBuilder.BuildRuleFromString(eventRule)
	assert.NoError(t, err)
	eng := engine.NewGengine()
	err = eng.ExecuteConcurrent(ruleBuilder)
	assert.NoError(t, err)
	t.Logf("temperature Event: [%s]\n", temperature.Event)
	t.Logf("water Event: [%s]\n", water.Event)
	t.Logf("smoke Event: [%s]\n", smoke.Event)
	for i := 0; i < 10; i++ {
		smoke.Value = int64(i % 3)
		err = eng.ExecuteConcurrent(ruleBuilder)
		assert.NoError(t, err)
		t.Logf("smoke Event: [%s]\n", smoke.Event)
	}
}

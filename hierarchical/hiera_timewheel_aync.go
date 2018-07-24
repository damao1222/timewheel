/**
 * Copyright (C) 2018, Xiongfa Li.
 * All right reserved.
 * @author xiongfa.li
 * @date 2018/7/12 
 * @time 14:00
 * @version V1.0
 * Description: 
 */

package hierarchical

import (
    "time"
    "github.com/xfali/timewheel"
    "github.com/xfali/timewheel/sync"
)

//Hierarchical Timing Wheels
type HieraTimeWheel struct {
    timeWheels [4] timewheel.TimeWheel
    tickTime   time.Duration
    stop       chan bool
}

func NewHieraTimeWheel(tickTime time.Duration, duration time.Duration) *HieraTimeWheel {
    tw := &HieraTimeWheel{}
    secondTick := false
    hour := duration / time.Hour
    if hour > 0 {
        secondTick = true
        wheel := sync.New(time.Hour, hour*time.Hour)
        tw.timeWheels[0] = wheel
    }

    minute := (duration % time.Hour) / time.Minute
    if hour > 0 {
        wheel := sync.New(time.Minute, time.Hour)
        wheel.Add(func() {
            tw.timeWheels[0].Tick(time.Hour)
        }, time.Hour, true)
        tw.timeWheels[1] = wheel
    } else {
        if minute > 0 {
            secondTick = true
            wheel := sync.New(time.Minute, minute*time.Minute)
            tw.timeWheels[1] = wheel
        }
    }

    second := (duration % time.Minute) / time.Second
    if minute > 0 {
        wheel := sync.New(time.Second, time.Minute)
        wheel.Add(func() {
            tw.timeWheels[1].Tick(time.Minute)
        }, time.Minute, true)
        tw.timeWheels[2] = wheel
    } else {
        if second > 0 {
            secondTick = true
            wheel := sync.New(time.Second, second*time.Second)
            tw.timeWheels[2] = wheel
        }
    }

    millisecond := (duration % time.Second) / time.Millisecond
    if secondTick {
        wheel := sync.New(tickTime, time.Second)
        wheel.Add(func() {
            tw.timeWheels[2].Tick(time.Second)
        }, time.Second, true)
        tw.timeWheels[3] = wheel
    } else {
        if millisecond > 0 {
            wheel := sync.New(tickTime, millisecond*time.Millisecond)
            tw.timeWheels[3] = wheel
        }
    }
    tw.tickTime = tickTime
    return tw
}

func (htw *HieraTimeWheel) Start() {
    go func() {
        now := time.Now()
        cur := now
        for {
            select {
            case <-htw.stop:
                return
            default:
                passTime := time.Since(now)
                if passTime < htw.tickTime {
                    time.Sleep(htw.tickTime - passTime)
                }
                cur = time.Now()
                htw.Tick(htw.tickTime)
                now = cur
            }
        }
    }()
}

func (htw *HieraTimeWheel) Stop() {
    close(htw.stop)
}

func (htw *HieraTimeWheel) Tick(duration time.Duration) {
    htw.timeWheels[3].Tick(duration)
}

func (htw *HieraTimeWheel) Add(callback timewheel.OnTimeout, expire time.Duration, repeat bool) (timewheel.CancelFunc, error) {
    return htw.addHour(callback, expire, repeat)
}

func (htw *HieraTimeWheel)addHour(callback timewheel.OnTimeout, expire time.Duration, repeat bool) (timewheel.CancelFunc, error) {
    hour := expire / time.Hour
    if hour > 0 {
        return htw.timeWheels[0].Add(func() {
            htw.addMinute(callback, expire, false)
        }, hour*time.Hour, repeat)
    } else {
        return htw.addMinute(callback, expire, repeat)
    }
}

func (htw *HieraTimeWheel)addMinute(callback timewheel.OnTimeout, expire time.Duration, repeat bool) (timewheel.CancelFunc, error) {
    minute := expire / time.Minute
    if minute > 0 {
        return htw.timeWheels[1].Add(func() {
            htw.addSecond(callback, expire, false)
        }, minute*time.Minute, repeat)
    } else {
        return htw.addSecond(callback, expire, repeat)
    }
}

func (htw *HieraTimeWheel)addSecond(callback timewheel.OnTimeout, expire time.Duration, repeat bool) (timewheel.CancelFunc, error) {
    second := expire / time.Second
    if second > 0 {
        return htw.timeWheels[0].Add(func() {
            htw.addMilliSecond(callback, expire, false)
        }, second*time.Second, repeat)
    } else {
        return htw.addMilliSecond(callback, expire, repeat)
    }
}

func (htw *HieraTimeWheel)addMilliSecond(callback timewheel.OnTimeout, expire time.Duration, repeat bool) (timewheel.CancelFunc, error) {
    millisecond := expire / time.Millisecond
    if millisecond > 0 {
        return htw.timeWheels[0].Add(callback, millisecond*time.Millisecond, repeat)
    } else {
        callback()
        return undo, nil
    }
}

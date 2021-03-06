// Copyright 2017 gf Author(https://gitee.com/johng/gf). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://gitee.com/johng/gf.

// 时间管理
package gtime

import (
    "time"
    "regexp"
    "strings"
    "strconv"
    "errors"
)

const (
    // 常用时间格式正则匹配，支持的标准时间格式：
    // "2017-12-14 04:51:34 +0805 LMT",
    // "2017-12-14 04:51:34 +0805 LMT",
    // "2006-01-02T15:04:05Z07:00",
    // "2014-01-17T01:19:15+08:00",
    // "2018-02-09T20:46:17.897Z",
    // "2018-02-09 20:46:17.897",
    // "2018-02-09T20:46:17Z",
    // "2018-02-09 20:46:17",
    // "2018-02-09",
    // 日期连接符号支持'-'或者'/'
    TIME_REAGEX_PATTERN = `(\d{2,4}[-/]\d{2}[-/]\d{2})[\sT]{0,1}(\d{0,2}:{0,1}\d{0,2}:{0,1}\d{0,2}){0,1}\.{0,1}(\d{0,9})([\sZ]{0,1})([\+-]{0,1})([:\d]*)`
)

var (
    // 使用正则判断会比直接使用ParseInLocation挨个轮训判断要快很多
    timeRegex, _   = regexp.Compile(TIME_REAGEX_PATTERN)
)

// 类似与js中的SetTimeout，一段时间后执行回调函数
func SetTimeout(t time.Duration, callback func()) {
    go func() {
        time.Sleep(t)
        callback()
    }()
}

// 类似与js中的SetInterval，每隔一段时间后执行回调函数，当回调函数返回true，那么继续执行，否则终止执行，该方法是异步的
// 注意：由于采用的是循环而不是递归操作，因此间隔时间将会以上一次回调函数执行完成的时间来计算
func SetInterval(t time.Duration, callback func() bool) {
    go func() {
        for {
            time.Sleep(t)
            if !callback() {
                break
            }
        }
    }()
}

// 设置当前进程全局的默认时区，如: Asia/Shanghai
func SetTimeZone(zone string) error {
    location, err := time.LoadLocation(zone)
    if err == nil {
        time.Local = location
    }
    return err
}

// 获取当前的纳秒数
func Nanosecond() int64 {
    return time.Now().UnixNano()
}

// 获取当前的微秒数
func Microsecond() int64 {
    return time.Now().UnixNano()/1e3
}

// 获取当前的毫秒数
func Millisecond() int64 {
    return time.Now().UnixNano()/1e6
}

// 获取当前的秒数(时间戳)
func Second() int64 {
    return time.Now().UnixNano()/1e9
}

// 获得当前的日期(例如：2006-01-02)
func Date() string {
    return time.Now().Format("2006-01-02")
}

// 获得当前的时间(例如：2006-01-02 15:04:05)
func Datetime() string {
    return time.Now().Format("2006-01-02 15:04:05")
}

// 字符串转换为时间对象
func StrToTime(str string) (time.Time, error) {
    var result time.Time
    var local  = time.Local
    if match := timeRegex.FindStringSubmatch(str); len(match) > 0 {
        var year, month, day, hour, min, sec, nsec int
        var array []string
        // 日期(支持'-'或'/'连接符号)
        array = strings.Split(match[1], "-")
        if len(array) < 3 {
            array = strings.Split(match[1], "/")
        }
        if len(array) >= 3 {
            // 年是否为缩写，如果是，那么需要补上前缀
            year, _  = strconv.Atoi(array[0])
            if year < 100 {
                year = int(time.Now().Year()/100)*100 + year
            }
            month, _ = strconv.Atoi(array[1])
            day, _   = strconv.Atoi(array[2])
        }
        // 时间
        if len(match[2]) > 0 {
            array   = strings.Split(match[2], ":")
            hour, _ = strconv.Atoi(array[0])
            if len(array) >= 2 {
                min, _ = strconv.Atoi(array[1])
            }
            if len(array) >= 3 {
                sec, _ = strconv.Atoi(array[2])
            }
        }
        // 纳秒，检查并执行位补齐
        if len(match[3]) > 0 {
            nsec, _   = strconv.Atoi(match[3])
            for i := 0; i < 9 - len(match[3]); i++ {
                nsec *= 10
            }
        }
        // 如果字符串中有时区信息(具体时间信息)，那么执行时区转换，将时区转成UTC
        if match[4] != "" && match[6] == "" {
            match[6] = "000000"
        }
        // 如果offset有值优先处理offset，否则处理后面的时区名称
        if match[6] != "" {
            zone := strings.Replace(match[6], ":", "", -1)
            zone  = strings.TrimLeft(zone, "+-")
            zone += strings.Repeat("0", 6 - len(zone))
            h, _ := strconv.Atoi(zone[0 : 2])
            m, _ := strconv.Atoi(zone[2 : 4])
            s, _ := strconv.Atoi(zone[4 : 6])
            // 判断字符串输入的时区是否和当前程序时区相等(使用offset判断)，不相等则将对象统一转换为UTC时区
            // 当前程序时区Offset(秒)
            _, localOffset := time.Now().Zone()
            if (h * 3600 + m * 60 + s) != localOffset {
                local = time.UTC
                // UTC时差转换
                operation := match[5]
                if operation != "+" && operation != "-" {
                    operation = "-"
                }
                switch operation {
                    case "+":
                        if h > 0 {
                            hour -= h
                        }
                        if m > 0 {
                            min  -= m
                        }
                        if s > 0 {
                            sec  -= s
                        }
                    case "-":
                        if h > 0 {
                            hour += h
                        }
                        if m > 0 {
                            min  += m
                        }
                        if s > 0 {
                            sec  += s
                        }
                }
            }
        }
        // 生成UTC时间对象
        result = time.Date(year, time.Month(month), day, hour, min, sec, nsec, local)
        return result, nil
    }
    return result, errors.New("unsupported time format")
}

// 时区转换
func ConvertZone(strTime string, toZone string, fromZone...string) (time.Time, error) {
   t, err := StrToTime(strTime)
   if err != nil {
       return time.Time{}, err
   }
   if len(fromZone) > 0 {
       if l, err := time.LoadLocation(fromZone[0]); err != nil {
           return time.Time{}, err
       } else {
           t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), l)
       }
   }
    if l, err := time.LoadLocation(toZone); err != nil {
        return time.Time{}, err
    } else {
        return t.In(l), nil
    }
}

// 字符串转换为时间对象，指定字符串时间格式，format格式形如：Y-m-d H:i:s
func StrToTimeFormat(str string, format string) (time.Time, error) {
    return StrToTimeLayout(str, formatToStdLayout(format))
}
// 字符串转换为时间对象，通过标准库layout格式进行解析，layout格式形如：2006-01-02 15:04:05
func StrToTimeLayout(str string, layout string) (time.Time, error) {
    if t, err := time.ParseInLocation(layout, str, time.Local); err == nil {
        return t, nil
    } else {
        return time.Time{}, err
    }
}

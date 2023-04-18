package main

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type MemStatItem struct {
	Megabytes float64
	Command   string
	Fullline  string
}

var spaces = regexp.MustCompile(`\s+`)

func MemStats(contains string) ([]MemStatItem, error) {
	cmd := exec.Command("ps", "vax")
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	b, err := cmd.Output()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	stderrStr := stderr.String()
	if len(stderrStr) > 0 {
		logrus.Warnf("stderrStr: %+v", stderrStr)
	}
	lines := strings.Split(string(b), "\n")
	var memStatItemsHeaviestFirst []MemStatItem
	for _, l := range lines {
		if len(contains) > 0 && !ContainsCaseInsesitive(l, contains) {
			continue
		}
		l = strings.TrimSpace(l)
		columns := spaces.Split(l, -1)
		const (
			minColumns         = 10
			memoryInUseColumn  = 7
			commandStartColumn = 9
		)
		if len(columns) < minColumns {
			return nil, errors.Errorf("not enough columns")
		}
		memStr := strings.TrimSpace(columns[memoryInUseColumn])
		memKiloBytes, errA := strconv.Atoi(memStr)
		if errA != nil {
			return nil, errors.Wrapf(errA, "for line %+v and columns : %+v", l, strings.Join(columns, "|"))
		}
		megaBytes := float64(memKiloBytes) / 1024
		command := strings.Join(columns[commandStartColumn:], " ")
		msi := MemStatItem{Fullline: l, Command: command, Megabytes: megaBytes}
		memStatItemsHeaviestFirst = append(memStatItemsHeaviestFirst, msi)
	}
	sort.Slice(memStatItemsHeaviestFirst, func(i, j int) bool {
		return memStatItemsHeaviestFirst[i].Megabytes > memStatItemsHeaviestFirst[j].Megabytes
	})
	return memStatItemsHeaviestFirst, nil
}

func ContainsCaseInsesitive(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}


func LogChromeMem() {
	period := 10 * time.Second
	go func() {
		for {
			if err := Mem(); err != nil {
				logrus.Errorf("couldn't print mem chrome stats: %+v", err)
			}
			time.Sleep(period)
		}
	}()
}


func Mem() error {
	items, err := MemStats("chrome")
	if err != nil {
		return errors.WithStack(err)
	}
	var sumMegaBytes float64
	for _, item := range items {
		sumMegaBytes += item.Megabytes
	}
	var theHeaviest string
	if len(items) > 0 {
		heaviest := items[0]
		theHeaviest = fmt.Sprintf("%+v MB, command: '%+v', full line: '%+v'", heaviest.Megabytes, heaviest.Command, heaviest.Fullline)
	}
	logrus.Infof("chrome total megabytes: %+v, items: %+v, heaviest: %+v", sumMegaBytes, len(items), theHeaviest)
	return nil
}


func OpenFiles() (int, error) {
	//lsof -Fn | sort | uniq | wc -l
	cmd := exec.Command("lsof", "-Fn")
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	b, err := cmd.Output()
	if err != nil {
		return 0, errors.WithStack(err)
	}
	stderrStr := stderr.String()
	if len(stderrStr) > 0 {
		logrus.Warnf("stderrStr: %+v", stderrStr)
	}
	lines := strings.Split(string(b), "\n")
	unique := make(map[string]struct{})
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		unique[l] = struct{}{}
	}
	return len(unique), nil
}

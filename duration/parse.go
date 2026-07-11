// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package duration

import (
	"fmt"
	"strconv"
	"strings"
)

// parseMode is the ISO 8601 parser state.
type parseMode int

const (
	modeStart parseMode = iota
	modeSign
	modeDate
	modeTime
)

// Parse parses an ISO 8601 duration string into a Duration, mirroring
// ActiveSupport::Duration.parse / ISO8601Parser including its validation rules:
// weeks may not mix with other date parts, an empty T marker is rejected, and
// only the last non-zero part may be fractional.
func Parse(s string) (Duration, error) {
	parts, order, err := parseISO8601(s)
	if err != nil {
		return Duration{}, err
	}
	ordered := make([]Part, len(order))
	var total float64
	for i, u := range order {
		ordered[i] = Part{u, parts[u]}
		total += parts[u] * secondsPerUnit[u]
	}
	return newDuration(total, ordered), nil
}

func parseErr(s, reason string) error {
	msg := fmt.Sprintf("invalid ISO 8601 duration: %q", s)
	if reason != "" {
		msg += " " + reason
	}
	return fmt.Errorf("%w: %s", ErrParse, msg)
}

var dateToPart = map[byte]Unit{'Y': UnitYears, 'M': UnitMonths, 'W': UnitWeeks, 'D': UnitDays}
var timeToPart = map[byte]Unit{'H': UnitHours, 'M': UnitMinutes, 'S': UnitSeconds}

// parseISO8601 runs the scanner state machine and returns the parts plus their
// insertion order.
func parseISO8601(s string) (map[Unit]float64, []Unit, error) {
	parts := map[Unit]float64{}
	var order []Unit
	sign := 1.0
	mode := modeStart
	pos := 0

	for pos < len(s) {
		switch mode {
		case modeStart:
			if s[pos] == '-' {
				sign = -1
				pos++
			} else if s[pos] == '+' {
				pos++
			}
			mode = modeSign
		case modeSign:
			if pos < len(s) && s[pos] == 'P' {
				pos++
				mode = modeDate
			} else {
				return nil, nil, parseErr(s, "")
			}
		case modeDate:
			if pos < len(s) && s[pos] == 'T' {
				pos++
				mode = modeTime
				continue
			}
			num, desig, next, ok := scanComponent(s, pos, "YMWD")
			if !ok {
				return nil, nil, parseErr(s, "")
			}
			u := dateToPart[desig]
			setPart(parts, &order, u, num*sign)
			pos = next
		case modeTime:
			num, desig, next, ok := scanComponent(s, pos, "HMS")
			if !ok {
				return nil, nil, parseErr(s, "")
			}
			u := timeToPart[desig]
			setPart(parts, &order, u, num*sign)
			pos = next
		}
	}

	if err := validateParse(s, parts, order, mode); err != nil {
		return nil, nil, err
	}
	return parts, order, nil
}

func setPart(parts map[Unit]float64, order *[]Unit, u Unit, v float64) {
	if _, ok := parts[u]; !ok {
		*order = append(*order, u)
	}
	parts[u] = v
}

// scanComponent matches (-?\d+(?:[.,]\d+)?)(designator) at pos, where the
// designator is one of allowed. It returns the value, the designator byte and
// the position past the designator.
func scanComponent(s string, pos int, allowed string) (float64, byte, int, bool) {
	i := pos
	if i < len(s) && s[i] == '-' {
		i++
	}
	digitStart := i
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == digitStart {
		return 0, 0, 0, false
	}
	if i < len(s) && (s[i] == '.' || s[i] == ',') {
		i++
		fracStart := i
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		if i == fracStart {
			return 0, 0, 0, false
		}
	}
	if i >= len(s) || strings.IndexByte(allowed, s[i]) < 0 {
		return 0, 0, 0, false
	}
	desig := s[i]
	numStr := strings.Replace(s[pos:i], ",", ".", 1)
	// numStr is guaranteed syntactically valid by the scan above.
	v, _ := strconv.ParseFloat(numStr, 64)
	return v, desig, i + 1, true
}

func validateParse(s string, parts map[Unit]float64, order []Unit, mode parseMode) error {
	if len(parts) == 0 {
		return parseErr(s, "is empty duration")
	}
	if _, hasWeeks := parts[UnitWeeks]; hasWeeks {
		_, y := parts[UnitYears]
		_, m := parts[UnitMonths]
		_, d := parts[UnitDays]
		if y || m || d {
			return parseErr(s, "mixing weeks with other date parts not allowed")
		}
	}
	if mode == modeTime {
		_, h := parts[UnitHours]
		_, mi := parts[UnitMinutes]
		_, se := parts[UnitSeconds]
		if !h && !mi && !se {
			return parseErr(s, "time part marker is present but time part is empty")
		}
	}
	// Only the last non-zero part may be fractional.
	var lastNonZero = -1
	for i, u := range order {
		if parts[u] != 0 {
			lastNonZero = i
		}
	}
	for i, u := range order {
		v := parts[u]
		if v == 0 || v == float64(int64(v)) {
			continue
		}
		if i != lastNonZero {
			return parseErr(s, "(only last part can be fractional)")
		}
	}
	return nil
}

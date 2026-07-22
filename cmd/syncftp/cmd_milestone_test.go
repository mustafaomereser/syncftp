package main

import (
	"testing"
	"time"
)

func TestParseMilestoneDate(t *testing.T) {
	now := time.Now()

	t.Run("iso", func(t *testing.T) {
		got, err := parseMilestoneDate("2026-07-20 14:30")
		if err != nil {
			t.Fatal(err)
		}
		want := time.Date(2026, 7, 20, 14, 30, 0, 0, time.Local)
		if !got.Equal(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("dotted", func(t *testing.T) {
		got, err := parseMilestoneDate("20.07.2026")
		if err != nil {
			t.Fatal(err)
		}
		want := time.Date(2026, 7, 20, 0, 0, 0, 0, time.Local)
		if !got.Equal(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("day-month-current-year", func(t *testing.T) {
		got, err := parseMilestoneDate("20.07")
		if err != nil {
			t.Fatal(err)
		}
		want := time.Date(now.Year(), 7, 20, 0, 0, 0, 0, time.Local)
		if !got.Equal(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("now", func(t *testing.T) {
		for _, s := range []string{"now", "şimdi", "simdi", "ŞİMDİ"} {
			got, err := parseMilestoneDate(s)
			if err != nil {
				t.Fatalf("%q: %v", s, err)
			}
			if d := time.Since(got); d < 0 || d > time.Minute {
				t.Fatalf("%q: got %v, expected ~now", s, got)
			}
		}
	})

	t.Run("today", func(t *testing.T) {
		for _, s := range []string{"bugün", "bugun", "today", "BUGÜN"} {
			got, err := parseMilestoneDate(s)
			if err != nil {
				t.Fatalf("%q: %v", s, err)
			}
			want := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
			if !got.Equal(want) {
				t.Fatalf("%q: got %v, want %v", s, got, want)
			}
		}
	})

	t.Run("yesterday", func(t *testing.T) {
		got, err := parseMilestoneDate("dün")
		if err != nil {
			t.Fatal(err)
		}
		y := now.AddDate(0, 0, -1)
		want := time.Date(y.Year(), y.Month(), y.Day(), 0, 0, 0, 0, time.Local)
		if !got.Equal(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("relative", func(t *testing.T) {
		got, err := parseMilestoneDate("3d")
		if err != nil {
			t.Fatal(err)
		}
		if d := time.Now().AddDate(0, 0, -3).Sub(got); d < -time.Minute || d > time.Minute {
			t.Fatalf("3d: got %v, expected ~3 days ago", got)
		}
		got, err = parseMilestoneDate("5h")
		if err != nil {
			t.Fatal(err)
		}
		if d := time.Now().Add(-5 * time.Hour).Sub(got); d < -time.Minute || d > time.Minute {
			t.Fatalf("5h: got %v, expected ~5 hours ago", got)
		}
	})

	t.Run("time-only", func(t *testing.T) {
		got, err := parseMilestoneDate("14:30")
		if err != nil {
			t.Fatal(err)
		}
		want := time.Date(now.Year(), now.Month(), now.Day(), 14, 30, 0, 0, time.Local)
		if !got.Equal(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for _, s := range []string{"", "abc", "99.99", "2026-13-40"} {
			if _, err := parseMilestoneDate(s); err == nil {
				t.Fatalf("%q: hata bekleniyordu", s)
			}
		}
	})
}

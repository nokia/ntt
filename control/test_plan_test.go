package control_test

import (
	"testing"

	"github.com/nokia/ntt/control"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/project"
	"github.com/stretchr/testify/assert"
)

func TestTestPlan(t *testing.T) {
	t.Parallel()
	t.Run("not nil", func(t *testing.T) {
		assert.Panics(t, func() { control.NewTestPlan(nil) })
	})

	t.Run("empty", func(t *testing.T) {
		tp, _ := control.NewTestPlan(&project.Config{})
		assert.NotNil(t, tp)
		assert.Nil(t, tp.Next())
	})

	t.Run("unknown test", func(t *testing.T) {
		tp, _ := control.NewTestPlan(&project.Config{})
		assert.NotNil(t, tp)
		err := tp.Add("unknown")
		assert.NotNil(t, err)
		assert.Nil(t, tp.Next())
	})

	t.Run("order of tests and control functions", func(t *testing.T) {
		fs.SetContent("test://file1.ttcn3", []byte(`module m1 { }`))
		fs.SetContent("test://file2.ttcn3", []byte(`
			module m1 {
				function f1() {}
				function @control f2() {}
				function @control f3() {}
				function @control f3() {}
				testcase tc1() {}
				control {}
			}
			control {}
			module m2 {
				testcase tc1() {}
				testcase tc2() {}
				testcase tc3() {}
		`))
		fs.SetContent("test://file3.ttcn3", []byte(`testcase tc4() {}`))
		fs.SetContent("test://file4.ttcn3", []byte(`
			module m3 {
				control {}
			}`))
		fs.SetContent("test://file5.ttcn3", []byte(``))
		tp, err := control.NewTestPlan(&project.Config{
			Manifest: project.Manifest{
				Sources: []string{
					"test://file1.ttcn3",
					"test://file2.ttcn3",
					"test://file3.ttcn3",
					"test://file4.ttcn3",
					"test://file5.ttcn3",
				}}})
		assert.Nil(t, err)
		if tp == nil {
			t.Fatal("test plan is nil")
		}
		assert.Equal(t, []string{
			"m1.tc1",
			"m2.tc1",
			"m2.tc2",
			"m2.tc3",
			"tc4",
		}, tp.Tests)
		assert.Equal(t, []string{
			"m1.f2",
			"m1.f3",
			"m1.f3", // duplicates are allowed
			"m1.control",
			"control", // module name is optional
			"m3.control",
		}, tp.Controls)

	})
}

package gorp_test

import (
	"github.com/developer-kikikaikai/go_routine_pool"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	WaitSec     = 3
	NumOfWorker = 4
)

/* Structure to test running thread */
type Result struct {
	sync.Mutex
	result int
}

//add result
func (r *Result) Add(value int) {
	r.Lock()
	r.result += value
	r.Unlock()
}
func (r *Result) Result() int {
	return r.result
}

type testRunningActor struct {
	value  int
	result *Result
	wg     *sync.WaitGroup
}

func (actor *testRunningActor) Do() {
	time.Sleep(WaitSec * time.Second)
	actor.result.Add(actor.value)
	actor.wg.Done()
}

type testFuncCloseActor struct {
	in int
}

func (actor *testFuncCloseActor) Do() {
	actor.in++
}

func TestRunningWorker(t *testing.T) {
	var worker gorp.GoRoutineWorker
	//Run worker
	Convey("Check Worker is running normally", t, func() {
		_worker, err := gorp.RunWorker(NumOfWorker)
		So(err, ShouldBeNil)
		worker = _worker
		result := Result{}
		expect := 0
		wg := &sync.WaitGroup{}
		begin := time.Now()
		for i := 0; i < NumOfWorker; i++ {
			value := 1 << i
			//Get time before
			wg.Add(1)
			worker.Go(&testRunningActor{value, &result, wg})
			//check this function is not blocked
			end := time.Now()
			expectTimeLess := begin.Add(time.Second)
			So(end.Before(expectTimeLess), ShouldBeTrue)

			//update expect value
			expect += value
		}
		wg.Wait()
		//Check result
		So(result.Result(), ShouldEqual, expect)
		end := time.Now()
		//result : WaitSec < do all action < WaitSec + 1 sec
		expectTimeGreater := begin.Add(WaitSec * time.Second)
		So(end.After(expectTimeGreater), ShouldBeTrue)
		expectTimeLess := begin.Add((WaitSec + 1) * time.Second)
		So(end.Before(expectTimeLess), ShouldBeTrue)

		//Check stop
		worker.Stop()
		actor := testFuncCloseActor{}
		value := actor.in
		worker.Go(&actor)
		//actor won't update actor.in param because there is no worker routine
		time.Sleep(time.Millisecond * 200)
		//check result
		So(actor.in, ShouldEqual, value)
	})

	Convey("Check running worker num", t, func() {
		_worker, err := gorp.RunWorker(NumOfWorker)
		So(err, ShouldBeNil)
		worker = _worker
		result := Result{}
		expect := 0
		wg := &sync.WaitGroup{}
		begin := time.Now()
		for i := 0; i < NumOfWorker+1; i++ {
			value := 1 << i
			//Get time before
			wg.Add(1)
			worker.Go(&testRunningActor{value, &result, wg})
			//check this function is not blocked
			end := time.Now()
			expectTimeLess := begin.Add(time.Second)
			So(end.Before(expectTimeLess), ShouldBeTrue)
			//update expect value
			expect += value
		}

		wg.Wait()
		//Check result
		So(result.Result(), ShouldEqual, expect)
		end := time.Now()
		//result : WaitSec * 2 < do all action < WaitSec * 2 + 1 sec
		// Because 1 action was blocked
		expectTimeGreater := begin.Add(WaitSec * 2 * time.Second)
		So(end.After(expectTimeGreater), ShouldBeTrue)
		expectTimeLess := begin.Add(((WaitSec * 2) + 1) * time.Second)
		So(end.Before(expectTimeLess), ShouldBeTrue)
	})
}
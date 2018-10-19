package smoke_tests

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Run a job", func() {

	BeforeEach(func() {
		Eventually(func() *gexec.Session {
			login := spawnFlyLogin()
			<-login.Exited
			return login
		}, 2*time.Minute, time.Second).Should(gexec.Exit(0))

		fly("set-pipeline", "-p", "trigger-pipeline", "-c", "fixtures/trigger-pipeline.yml", "-n")
		fly("unpause-pipeline", "-p", "trigger-pipeline")
	})

	AfterEach(func() {
		fly("destroy-pipeline", "-p", "trigger-pipeline", "-n")
	})

	It("can run a job and connect to the internet", func() {
		fly("trigger-job", "--job", "trigger-pipeline/do-something")

		watch := waitForBuildAndWatch("trigger-pipeline/do-something")
		Eventually(watch).Should(gbytes.Say("some-resource/input"))
	})
})

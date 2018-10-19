package smoke_tests

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Set a pipeline", func() {

	BeforeEach(func() {
		Eventually(func() *gexec.Session {
			login := spawnFlyLogin()
			<-login.Exited
			return login
		}, 2*time.Minute, time.Second).Should(gexec.Exit(0))
	})

	AfterEach(func() {
		fly("destroy-pipeline", "-p", "tiny-pipeline", "-n")
	})

	It("can set a pipeline", func() {
		fly("set-pipeline", "-p", "tiny-pipeline", "-c", "fixtures/tiny-pipeline.yml", "-n")
	})
})

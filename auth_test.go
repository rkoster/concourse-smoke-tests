package smoke_tests

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Concourse Auth", func() {

	FIt("can log into the concourse", func() {
		Eventually(func() *gexec.Session {
			login := spawnFlyLogin()
			<-login.Exited
			return login
		}, 2*time.Minute, time.Second).Should(gexec.Exit(0))

	})
})

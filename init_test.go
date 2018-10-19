// Heavily inspired by:
// https://github.com/concourse/concourse/blob/master/testflight/suite_test.go

package smoke_tests

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const flyTarget = "st"

var (
	concourseUrl string
	username     string
	password     string

	flyPath           string
	skipSSLValidation bool
)

func TestSmoke(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "concourse-smoke-tests")
}

var _ = BeforeSuite(func() {
	concourseUrl = os.Getenv("CONCOURSE_URL")
	if concourseUrl == "" {
		Fail("CONCOURSE_URL is a required paramter")
	}

	username = os.Getenv("USERNAME")
	password = os.Getenv("PASSWORD")

	flyPath = os.Getenv("FLY_PATH")
	if flyPath == "" {
		flyPath = "fly"
	}

	if os.Getenv("FLY_SKIP_SSL") == "true" {
		skipSSLValidation = true
	}
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

func spawn(argc string, argv ...string) *gexec.Session {
	By("running: " + argc + " " + strings.Join(argv, " "))
	cmd := exec.Command(argc, argv...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	return session
}

func spawnFly(argv ...string) *gexec.Session {
	return spawn(flyPath, append([]string{"-t", flyTarget}, argv...)...)
}

func fly(argv ...string) *gexec.Session {
	sess := spawnFly(argv...)
	wait(sess)
	return sess
}

func wait(session *gexec.Session) {
	<-session.Exited
	Expect(session.ExitCode()).To(Equal(0))
}

func spawnFlyLogin(args ...string) *gexec.Session {
	extraArgs := []string{"login", "-c", concourseUrl, "-u", username, "-p", password}
	if skipSSLValidation {
		extraArgs = append(extraArgs, "-k")
	}

	return spawnFly(append(extraArgs, args...)...)
}

func waitForBuildAndWatch(jobName string, buildName ...string) *gexec.Session {
	args := []string{"watch", "-j", jobName}

	if len(buildName) > 0 {
		args = append(args, "-b", buildName[0])
	}

	keepPollingCheck := regexp.MustCompile("job has no builds|build not found|failed to get build")
	for {
		session := spawnFly(args...)
		<-session.Exited

		if session.ExitCode() == 1 {
			output := strings.TrimSpace(string(session.Err.Contents()))
			if keepPollingCheck.MatchString(output) {
				// build hasn't started yet; keep polling
				time.Sleep(time.Second)
				continue
			}
		}

		return session
	}
}

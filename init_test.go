package smoke_tests

import (
	"os"
	"os/exec"
	"strings"
	"testing"

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

func spawnFlyLogin(args ...string) *gexec.Session {
	extraArgs := []string{"login", "-c", concourseUrl, "-u", username, "-p", password}
	if skipSSLValidation {
		extraArgs = append(extraArgs, "-k")
	}

	return spawnFly(append(extraArgs, args...)...)
}

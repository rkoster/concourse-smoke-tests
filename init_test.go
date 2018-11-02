// Heavily inspired by:
// https://github.com/concourse/concourse/blob/master/testflight/suite_test.go

package smoke_tests

import (
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	yaml "gopkg.in/yaml.v2"
)

const flyTarget = "st"

var source ConcourseSource

func TestSmoke(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "concourse-smoke-tests")
}

type LocalUser struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type ConcourseSource struct {
	ExternalURL       string    `yaml:"external_url"`
	LocalUser         LocalUser `yaml:"local_user"`
	SkipSSLValidation bool      `yaml:"skip_ssl_validation"`
	FlyPath           string    `yaml:"fly_path"`
}

var _ = BeforeSuite(func() {
	sourceFile := os.Getenv("CONCOURSE_SOURCE_FILE")
	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		source = ConcourseSource{}

		sourceYml, err := ioutil.ReadFile(sourceFile)
		Expect(err).ToNot(HaveOccurred())

		err = yaml.Unmarshal(sourceYml, &source)
		Expect(err).ToNot(HaveOccurred())

	} else {
		source = ConcourseSource{
			ExternalURL: os.Getenv("CONCOURSE_URL"),
			LocalUser: LocalUser{
				Username: os.Getenv("CONCOURSE_USERNAME"),
				Password: os.Getenv("CONCOURSE_PASSWORD"),
			},
			FlyPath: os.Getenv("FLY_PATH"),
		}
	}
	skipSSLValidation, err := strconv.ParseBool(os.Getenv("FLY_SKIP_SSL"))
	if err != nil {
		skipSSLValidation = false
	}
	source.SkipSSLValidation = skipSSLValidation
	if source.ExternalURL == "" {
		Fail("CONCOURSE_URL is a required paramter")
	}
	if source.FlyPath == "" {
		source.FlyPath = "fly"
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
	return spawn(source.FlyPath, append([]string{"-t", flyTarget}, argv...)...)
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
	extraArgs := []string{"login", "-c", source.ExternalURL, "-u", source.LocalUser.Username, "-p", source.LocalUser.Password}
	if source.SkipSSLValidation {
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

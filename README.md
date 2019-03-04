## Concourse smoke tests.

Very basic smoke tests to make sure the instance you just spun up can run a job. 

Note: It only can authenticate against the built in auth, next step will be to make it work with an oAuth flow. 

Usage:

```
pushd concourse-smoke-tests
  export CONCOURSE_URL=https://my-concourse.example.com
  export CONCOURSE_USERNAME=my-username
  export CONCOURSE_PASSWORD=my-password
  ginkgo -p .
popd

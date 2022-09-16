package dockerClient

import (
	"fmt"
	"os"
)

func createDockerFile() string {

	uid := os.Getuid()

	return fmt.Sprintf(`FROM aquasec/trivy:latest
RUN adduser --system -D -h /root -u %d trivy
USER trivy
ENTRYPOINT ["trivy"]
`, uid)

}

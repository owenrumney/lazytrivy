package docker

import (
	"fmt"
	"os"
)

func createDockerFile() string {

	uid := os.Getuid()

	return fmt.Sprintf(`FROM aquasec/trivy:latest
RUN adduser -D -h /root -u %d trivy
RUN chown -R trivy /root/.cache/
USER trivy
ENTRYPOINT ["trivy"]
`, uid)

}

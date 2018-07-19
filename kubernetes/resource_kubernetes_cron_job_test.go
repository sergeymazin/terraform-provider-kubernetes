package kubernetes

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"k8s.io/api/batch/v1beta1"
)

func TestAccKubernetesCronJob_basic(t *testing.T) {
	var conf v1beta1.CronJob
	name := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "kubernetes_cron_job.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckKubernetesCronJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesCronJobConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesCronJobExists("kubernetes_cron_job.test", &conf),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "metadata.0.name", name),
					resource.TestCheckResourceAttrSet("kubernetes_cron_job.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("kubernetes_cron_job.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("kubernetes_cron_job.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("kubernetes_cron_job.test", "metadata.0.uid"),
					resource.TestCheckResourceAttrSet("kubernetes_cron_job.test", "spec.0.schedule"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.job_template.0.spec.0.parallelism", "1"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.job_template.0.spec.0.backoff_limit", "2"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.job_template.0.spec.0.template.0.spec.0.container.0.name", "hello"),
				),
			},
			{
				Config: testAccKubernetesCronJobConfig_modified(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesCronJobExists("kubernetes_cron_job.test", &conf),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "metadata.0.name", name),
					resource.TestCheckResourceAttrSet("kubernetes_cron_job.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("kubernetes_cron_job.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("kubernetes_cron_job.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("kubernetes_cron_job.test", "metadata.0.uid"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.job_template.0.spec.0.parallelism", "2"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.job_template.0.spec.0.backoff_limit", "0"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.job_template.0.spec.0.template.0.spec.0.container.0.name", "hello"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.job_template.0.spec.0.template.0.metadata.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.job_template.0.spec.0.template.0.metadata.0.labels.%", "2"),
				),
			},
		},
	})
}

func TestAccKubernetesCronJob_extra(t *testing.T) {
	var conf v1beta1.CronJob
	name := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "kubernetes_cron_job.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckKubernetesCronJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesCronJobConfig_extra(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesCronJobExists("kubernetes_cron_job.test", &conf),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "metadata.0.name", name),
					resource.TestCheckResourceAttrSet("kubernetes_cron_job.test", "spec.0.schedule"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.concurrency_policy", "Forbid"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.successful_jobs_history_limit", "10"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.failed_jobs_history_limit", "10"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.starting_deadline_seconds", "60"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.job_template.0.spec.0.backoff_limit", "2"),
				),
			},
			{
				Config: testAccKubernetesCronJobConfig_extraModified(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesCronJobExists("kubernetes_cron_job.test", &conf),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "metadata.0.name", name),
					resource.TestCheckResourceAttrSet("kubernetes_cron_job.test", "spec.0.schedule"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.concurrency_policy", "Forbid"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.successful_jobs_history_limit", "2"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.failed_jobs_history_limit", "2"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.starting_deadline_seconds", "120"),
					resource.TestCheckResourceAttr("kubernetes_cron_job.test", "spec.0.job_template.0.spec.0.backoff_limit", "3"),
				),
			},
		},
	})
}

func testAccCheckKubernetesCronJobDestroy(s *terraform.State) error {
	kp := testAccProvider.Meta().(*kubernetesProvider)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "kubernetes_cron_job" {
			continue
		}

		namespace, name, err := idParts(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := readCronJob(kp, namespace, name)
		if err == nil {
			if resp.Name == rs.Primary.ID {
				return fmt.Errorf("CronJob still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckKubernetesCronJobExists(n string, obj *v1beta1.CronJob) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		kp := testAccProvider.Meta().(*kubernetesProvider)

		namespace, name, err := idParts(rs.Primary.ID)
		if err != nil {
			return err
		}

		out, err := readCronJob(kp, namespace, name)
		if err != nil {
			return err
		}

		*obj = *out
		return nil
	}
}

func testAccKubernetesCronJobConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "kubernetes_cron_job" "test" {
	metadata {
		name = "%s"
	}
	spec {
		schedule = "1 0 * * *"
		job_template {
			spec {
				backoff_limit = 2
				template {
					spec {
						container {
							name = "hello"
							image = "alpine"
							command = ["echo", "'hello'"]
						}
					}
				}
			}
		}
	}
}`, name)
}

func testAccKubernetesCronJobConfig_modified(name string) string {
	return fmt.Sprintf(`
resource "kubernetes_cron_job" "test" {
	metadata {
		name = "%s"
	}
	spec {
		schedule = "1 0 * * *"
		job_template {
			spec {
				parallelism = 2
				template {
					metadata {
						labels {
							foo = "bar"
							baz = "foo"
						}
					}
					spec {
						container {
							name = "hello"
							image = "alpine"
							command = ["echo", "'abcdef'"]
						}
					}
				}
			}
		}
	}
}`, name)
}

func testAccKubernetesCronJobConfig_extra(name string) string {
	return fmt.Sprintf(`
resource "kubernetes_cron_job" "test" {
	metadata {
		name = "%s"
	}
	spec {
		schedule = "1 0 * * *"
    	concurrency_policy            = "Forbid"
    	successful_jobs_history_limit = 10
    	failed_jobs_history_limit     = 10
    	starting_deadline_seconds     = 60
		job_template {
			spec {
				backoff_limit = 2
				template {
					spec {
						container {
							name = "hello"
							image = "alpine"
							command = ["echo", "'hello'"]
						}
					}
				}
			}
		}
	}
}`, name)
}

func testAccKubernetesCronJobConfig_extraModified(name string) string {
	return fmt.Sprintf(`
resource "kubernetes_cron_job" "test" {
	metadata {
		name = "%s"
	}
	spec {
		schedule = "1 0 * * *"
		concurrency_policy            = "Forbid"
		successful_jobs_history_limit = 2
		failed_jobs_history_limit     = 2
		starting_deadline_seconds     = 120
		job_template {
			spec {
				backoff_limit = 3
				template {
					spec {
						container {
							name = "hello"
							image = "alpine"
							command = ["echo", "'hello'"]
						}
					}
				}
			}
		}
	}
}`, name)
}

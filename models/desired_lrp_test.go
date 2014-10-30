package models_test

import (
	. "github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DesiredLRP", func() {
	var lrp DesiredLRP

	lrpPayload := `{
	  "process_guid": "some-guid",
	  "domain": "some-domain",
	  "root_fs": "docker:///docker.com/docker",
	  "instances": 1,
	  "stack": "some-stack",
	  "env":[
	    {
	      "name": "ENV_VAR_NAME",
	      "value": "some environment variable value"
	    }
	  ],
	  "actions": [
	    {
	      "action": "download",
	      "args": {
	        "from": "http://example.com",
	        "to": "/tmp/internet",
	        "cache_key": ""
	      }
	    }
	  ],
	  "disk_mb": 512,
	  "memory_mb": 1024,
	  "cpu_weight": 42,
	  "ports": [
	    {
	      "container_port": 5678,
	      "host_port": 1234
	    }
	  ],
	  "routes": [
	    "route-1",
	    "route-2"
	  ],
	  "log": {
	    "guid": "log-guid",
	    "source_name": "the cloud"
	  }
	}`

	BeforeEach(func() {
		lrp = DesiredLRP{
			Domain:      "some-domain",
			ProcessGuid: "some-guid",

			Instances:  1,
			Stack:      "some-stack",
			RootFSPath: "docker:///docker.com/docker",
			MemoryMB:   1024,
			DiskMB:     512,
			CPUWeight:  42,
			Routes:     []string{"route-1", "route-2"},
			Ports: []PortMapping{
				{HostPort: 1234, ContainerPort: 5678},
			},
			Log: LogConfig{
				Guid:       "log-guid",
				SourceName: "the cloud",
			},
			EnvironmentVariables: []EnvironmentVariable{
				{
					Name:  "ENV_VAR_NAME",
					Value: "some environment variable value",
				},
			},
			Actions: []ExecutorAction{
				{
					Action: DownloadAction{
						From: "http://example.com",
						To:   "/tmp/internet",
					},
				},
			},
		}
	})

	Describe("ToJSON", func() {
		It("should JSONify", func() {
			json := lrp.ToJSON()
			Ω(string(json)).Should(MatchJSON(lrpPayload))
		})
	})

	Describe("Validate", func() {
		var validationErr error

		JustBeforeEach(func() {
			validationErr = lrp.Validate()
		})

		Context("Domain", func() {
			BeforeEach(func() {
				lrp.Domain = ""
			})

			It("requires a domain", func() {
				Ω(validationErr).Should(HaveOccurred())
				Ω(validationErr.Error()).Should(ContainSubstring("domain"))
			})
		})

		Context("ProcessGuid", func() {
			BeforeEach(func() {
				lrp.ProcessGuid = ""
			})

			It("requires a process guid", func() {
				Ω(validationErr).Should(HaveOccurred())
				Ω(validationErr.Error()).Should(ContainSubstring("process_guid"))
			})
		})

		Context("Stack", func() {
			BeforeEach(func() {
				lrp.Stack = ""
			})

			It("requires a stack", func() {
				Ω(validationErr).Should(HaveOccurred())
				Ω(validationErr.Error()).Should(ContainSubstring("stack"))
			})
		})

		Context("Actions", func() {
			BeforeEach(func() {
				lrp.Actions = nil
			})

			It("requires actions", func() {
				Ω(validationErr).Should(HaveOccurred())
				Ω(validationErr.Error()).Should(ContainSubstring("actions"))
			})
		})
	})

	Describe("NewDesiredLRPFromJSON", func() {
		It("returns a LRP with correct fields", func() {
			decodedStartAuction, err := NewDesiredLRPFromJSON([]byte(lrpPayload))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(decodedStartAuction).Should(Equal(lrp))
		})

		Context("with an invalid payload", func() {
			It("returns the error", func() {
				decodedStartAuction, err := NewDesiredLRPFromJSON([]byte("aliens lol"))
				Ω(err).Should(HaveOccurred())

				Ω(decodedStartAuction).Should(BeZero())
			})
		})

		for field, payload := range map[string]string{
			"process_guid": `{
				"domain": "some-domain",
				"actions": [
					{"action":"download","args":{"from":"http://example.com","to":"/tmp/internet","cache_key":""}}
				],
				"stack": "some-stack"
			}`,
			"actions": `{
				"domain": "some-domain",
				"process_guid": "process_guid",
				"stack": "some-stack"
			}`,
			"stack": `{
				"domain": "some-domain",
				"process_guid": "process_guid",
				"actions": [
					{"action":"download","args":{"from":"http://example.com","to":"/tmp/internet","cache_key":""}}
				]
			}`,
			"domain": `{
				"stack": "some-stack",
				"process_guid": "process_guid",
				"actions": [
					{"action":"download","args":{"from":"http://example.com","to":"/tmp/internet","cache_key":""}}
				]
			}`,
		} {
			json := payload
			missingField := field

			Context("when the json is missing a "+missingField, func() {
				It("returns an error indicating so", func() {
					decodedStartAuction, err := NewDesiredLRPFromJSON([]byte(json))
					Ω(err).Should(HaveOccurred())
					Ω(err.Error()).Should(Equal("JSON has missing/invalid field: " + missingField))

					Ω(decodedStartAuction).Should(BeZero())
				})
			})
		}
	})
})

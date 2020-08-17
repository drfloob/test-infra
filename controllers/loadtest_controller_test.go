package controllers

import (
	"context"

	grpcv1 "github.com/grpc/test-infra/api/v1"
	"github.com/grpc/test-infra/pkg/defaults"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Test Environment", func() {
	It("supports creation of load tests", func() {
		err := k8sClient.Create(context.Background(), newLoadTest())
		Expect(err).ToNot(HaveOccurred())
	})
})

var _ = Describe("Pod Creation", func() {
	var loadtest *grpcv1.LoadTest

	BeforeEach(func() {
		loadtest = newLoadTest()
	})

	Describe("newClientPod", func() {
		var component *grpcv1.Component

		BeforeEach(func() {
			component = &loadtest.Spec.Clients[0].Component
		})

		It("sets component-name label", func() {
			name := "foo-bar-buzz"
			component.Name = &name

			pod, err := newClientPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())
			Expect(pod.Labels[defaults.ComponentNameLabel]).To(Equal(name))
		})

		It("sets loadtest-role label to client", func() {
			pod, err := newClientPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())
			Expect(pod.Labels[defaults.RoleLabel]).To(Equal(defaults.ClientRole))
		})

		It("sets loadtest label", func() {
			pod, err := newClientPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())
			Expect(pod.Labels[defaults.LoadTestLabel]).To(Equal(loadtest.Name))
		})

		It("sets node selector for appropriate pool", func() {
			customPool := "custom-pool"
			component.Pool = &customPool

			pod, err := newClientPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())
			Expect(pod.Spec.NodeSelector["pool"]).To(Equal(customPool))
		})

		It("sets clone init container", func() {
			cloneImage := "docker.pkg.github.com/grpc/test-infra/fake-image"
			component.Clone.Image = &cloneImage

			pod, err := newClientPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())

			expectedContainer := newCloneContainer(component.Clone)
			Expect(pod.Spec.InitContainers).To(ContainElement(expectedContainer))
		})

		It("sets build init container", func() {
			buildImage := "docker.pkg.github.com/grpc/test-infra/fake-image"

			build := new(grpcv1.Build)
			build.Image = &buildImage
			build.Command = []string{"bazel"}
			build.Args = []string{"build", "//target"}
			component.Build = build

			pod, err := newClientPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())

			expectedContainer := newBuildContainer(component.Build)
			Expect(pod.Spec.InitContainers).To(ContainElement(expectedContainer))
		})

		It("sets run container", func() {
			image := "golang:1.14"
			run := grpcv1.Run{
				Image:   &image,
				Command: []string{"go"},
				Args:    []string{"run", "main.go"},
			}
			component.Run = run

			pod, err := newClientPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())

			expectedContainer := newRunContainer(run)
			addDriverPort(&expectedContainer)
			Expect(pod.Spec.Containers).To(ContainElement(expectedContainer))
		})

		It("disables retries", func() {
			pod, err := newClientPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())
			Expect(pod.Spec.RestartPolicy).To(Equal(corev1.RestartPolicyNever))
		})

		It("exposes a driver port", func() {
			pod, err := newClientPod(loadtest, component)
			port := newContainerPort("driver", 10000)
			Expect(err).To(BeNil())
			Expect(pod.Spec.Containers[0].Ports).To(ContainElement(port))
		})
	})

	Describe("newDriverPod", func() {
		var component *grpcv1.Component

		BeforeEach(func() {
			component = &loadtest.Spec.Driver.Component
		})

		It("mounts GCP secrets", func() {
			// TODO: Add tests for mounting of GCP secrets
			Skip("complete this task when adding GCP secrets to pkg/defaults")
		})

		It("sets loadtest-role label to driver", func() {
			pod, err := newDriverPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())
			Expect(pod.Labels[defaults.RoleLabel]).To(Equal(defaults.DriverRole))
		})

		It("sets scenario environment variable", func() {
			// TODO: Add tests for configmap retrieval and scenario env variable
			Skip("add this when handling scenario ConfigMaps")
		})

		It("sets loadtest label", func() {
			pod, err := newDriverPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())
			Expect(pod.Labels[defaults.LoadTestLabel]).To(Equal(loadtest.Name))
		})

		It("sets component-name label", func() {
			name := "foo-bar-buzz"
			component.Name = &name

			pod, err := newDriverPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())
			Expect(pod.Labels[defaults.ComponentNameLabel]).To(Equal(name))
		})

		It("sets node selector for appropriate pool", func() {
			customPool := "custom-pool"
			component.Pool = &customPool

			pod, err := newDriverPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())
			Expect(pod.Spec.NodeSelector["pool"]).To(Equal(customPool))
		})

		It("sets clone init container", func() {
			cloneImage := "docker.pkg.github.com/grpc/test-infra/fake-image"
			component.Clone = new(grpcv1.Clone)
			component.Clone.Image = &cloneImage

			pod, err := newServerPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())

			expectedContainer := newCloneContainer(component.Clone)
			Expect(pod.Spec.InitContainers).To(ContainElement(expectedContainer))
		})

		It("sets build init container", func() {
			buildImage := "docker.pkg.github.com/grpc/test-infra/fake-image"

			build := new(grpcv1.Build)
			build.Image = &buildImage
			build.Command = []string{"bazel"}
			build.Args = []string{"build", "//target"}
			component.Build = build

			pod, err := newDriverPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())

			expectedContainer := newBuildContainer(component.Build)
			Expect(pod.Spec.InitContainers).To(ContainElement(expectedContainer))
		})

		It("sets run container", func() {
			image := "golang:1.14"
			run := grpcv1.Run{
				Image:   &image,
				Command: []string{"go"},
				Args:    []string{"run", "main.go"},
			}
			component.Run = run

			pod, err := newDriverPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())

			expectedContainer := newRunContainer(run)
			addDriverPort(&expectedContainer)
			Expect(pod.Spec.Containers).To(ContainElement(expectedContainer))
		})

		It("disables retries", func() {
			pod, err := newDriverPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())
			Expect(pod.Spec.RestartPolicy).To(Equal(corev1.RestartPolicyNever))
		})
	})

	Describe("newServerPod", func() {
		var component *grpcv1.Component

		BeforeEach(func() {
			component = &loadtest.Spec.Servers[0].Component
		})

		It("sets loadtest-role label to server", func() {
			pod, err := newServerPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())
			Expect(pod.Labels[defaults.RoleLabel]).To(Equal(defaults.ServerRole))
		})

		It("exposes a driver port", func() {
			pod, err := newServerPod(loadtest, component)
			port := newContainerPort("driver", 10000)
			Expect(err).To(BeNil())
			Expect(pod.Spec.Containers[0].Ports).To(ContainElement(port))
		})

		It("exposes a server port", func() {
			pod, err := newServerPod(loadtest, component)
			port := newContainerPort("server", 10010)
			Expect(err).To(BeNil())
			Expect(pod.Spec.Containers[0].Ports).To(ContainElement(port))
		})

		It("sets loadtest label", func() {
			pod, err := newServerPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())
			Expect(pod.Labels[defaults.LoadTestLabel]).To(Equal(loadtest.Name))
		})

		It("sets component-name label", func() {
			name := "foo-bar-buzz"
			component.Name = &name

			pod, err := newServerPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())
			Expect(pod.Labels[defaults.ComponentNameLabel]).To(Equal(name))
		})

		It("sets node selector for appropriate pool", func() {
			customPool := "custom-pool"
			component.Pool = &customPool

			pod, err := newServerPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())
			Expect(pod.Spec.NodeSelector["pool"]).To(Equal(customPool))
		})

		It("sets clone init container", func() {
			cloneImage := "docker.pkg.github.com/grpc/test-infra/fake-image"
			component.Clone.Image = &cloneImage

			pod, err := newServerPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())

			expectedContainer := newCloneContainer(component.Clone)
			Expect(pod.Spec.InitContainers).To(ContainElement(expectedContainer))
		})

		It("sets build init container", func() {
			buildImage := "docker.pkg.github.com/grpc/test-infra/fake-image"

			build := new(grpcv1.Build)
			build.Image = &buildImage
			build.Command = []string{"bazel"}
			build.Args = []string{"build", "//target"}
			component.Build = build

			pod, err := newServerPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())

			expectedContainer := newBuildContainer(component.Build)
			Expect(pod.Spec.InitContainers).To(ContainElement(expectedContainer))
		})

		It("sets run container", func() {
			image := "golang:1.14"
			run := grpcv1.Run{
				Image:   &image,
				Command: []string{"go"},
				Args:    []string{"run", "main.go"},
			}
			component.Run = run

			pod, err := newServerPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())

			expectedContainer := newRunContainer(run)
			addDriverPort(&expectedContainer)
			addServerPort(&expectedContainer)
			Expect(pod.Spec.Containers).To(ContainElement(expectedContainer))
		})

		It("disables retries", func() {
			pod, err := newServerPod(loadtest, component)
			Expect(err).ToNot(HaveOccurred())
			Expect(pod.Spec.RestartPolicy).To(Equal(corev1.RestartPolicyNever))
		})
	})

	Describe("newCloneContainer", func() {
		var clone *grpcv1.Clone

		BeforeEach(func() {
			image := "docker.pkg.github.com/grpc/test-infra/clone"
			repo := "https://github.com/grpc/test-infra.git"
			gitRef := "master"

			clone = &grpcv1.Clone{
				Image:  &image,
				Repo:   &repo,
				GitRef: &gitRef,
			}
		})

		It("sets the name of the container", func() {
			container := newCloneContainer(clone)
			Expect(container.Name).To(Equal(cloneInitContainer))
		})

		It("returns empty container given nil pointer", func() {
			clone = nil
			container := newCloneContainer(clone)
			Expect(container).To(Equal(corev1.Container{}))
		})

		It("sets clone image", func() {
			customImage := "debian:buster"
			clone.Image = &customImage

			container := newCloneContainer(clone)
			Expect(container.Image).To(Equal(customImage))
		})

		It("sets repo environment variable", func() {
			repo := "https://github.com/grpc/grpc.git"
			clone.Repo = &repo

			container := newCloneContainer(clone)
			Expect(container.Env).To(ContainElement(corev1.EnvVar{
				Name:  CloneRepoEnv,
				Value: repo,
			}))
		})

		It("sets git-ref environment variable", func() {
			gitRef := "master"
			clone.GitRef = &gitRef

			container := newCloneContainer(clone)
			Expect(container.Env).To(ContainElement(corev1.EnvVar{
				Name:  CloneGitRefEnv,
				Value: gitRef,
			}))
		})
	})

	Describe("newBuildContainer", func() {
		var build *grpcv1.Build

		BeforeEach(func() {
			image := "docker.pkg.github.com/grpc/test-infra/rust"

			build = &grpcv1.Build{
				Image:   &image,
				Command: nil,
				Args:    nil,
				Env:     nil,
			}
		})

		It("sets the name of the container", func() {
			container := newBuildContainer(build)
			Expect(container.Name).To(Equal(buildInitContainer))
		})

		It("returns empty container given nil pointer", func() {
			build = nil
			container := newBuildContainer(build)
			Expect(container).To(Equal(corev1.Container{}))
		})

		It("sets image", func() {
			customImage := "golang:latest"
			build.Image = &customImage

			container := newBuildContainer(build)
			Expect(container.Image).To(Equal(customImage))
		})

		It("sets command", func() {
			command := []string{"bazel"}
			build.Command = command

			container := newBuildContainer(build)
			Expect(container.Command).To(Equal(command))
		})

		It("sets args", func() {
			args := []string{"build", "//target"}
			build.Command = []string{"bazel"}
			build.Args = args

			container := newBuildContainer(build)
			Expect(container.Args).To(Equal(args))
		})

		It("sets environment variables", func() {
			env := []corev1.EnvVar{
				{Name: "EXPERIMENT", Value: "1"},
				{Name: "PROD", Value: "0"},
			}

			build.Env = env

			container := newBuildContainer(build)
			Expect(env[0]).To(BeElementOf(container.Env))
			Expect(env[1]).To(BeElementOf(container.Env))
		})
	})

	Describe("newRunContainer", func() {
		var run grpcv1.Run

		BeforeEach(func() {
			image := "docker.pkg.github.com/grpc/test-infra/fake-image"
			command := []string{"qps_worker"}

			run = grpcv1.Run{
				Image:   &image,
				Command: command,
			}
		})

		It("sets the name of the container", func() {
			container := newRunContainer(run)
			Expect(container.Name).To(Equal(runContainer))
		})

		It("sets image", func() {
			image := "golang:1.14"
			run.Image = &image

			container := newRunContainer(run)
			Expect(container.Image).To(Equal(image))
		})

		It("sets command", func() {
			command := []string{"go"}
			run.Command = command

			container := newRunContainer(run)
			Expect(container.Command).To(Equal(command))
		})

		It("sets args", func() {
			command := []string{"go"}
			args := []string{"run", "main.go"}
			run.Command = command
			run.Args = args

			container := newRunContainer(run)
			Expect(container.Args).To(Equal(args))
		})

		It("sets environment variables", func() {
			env := []corev1.EnvVar{
				{Name: "ENABLE_DEBUG", Value: "1"},
				{Name: "VERBOSE", Value: "1"},
			}

			run.Env = env

			container := newRunContainer(run)
			Expect(env[0]).To(BeElementOf(container.Env))
			Expect(env[1]).To(BeElementOf(container.Env))
		})
	})
})

var _ = Describe("CheckMissingPods", func() {

	var currentLoadTest = newLoadTestWithMultipleClientsAndServers() // nothing modify this --> okay
	var expectedReturnList []*grpcv1.Component                       //initialized every time after a test -->okay
	var allRunningPods = &corev1.PodList{Items: []corev1.Pod{}}
	var returnedList []*grpcv1.Component // it was a newlist generated form myfun everytime, address are different --> okay

	JustBeforeEach(func() {
		returnedList = CheckMissingPods(currentLoadTest, allRunningPods)
	})

	AfterEach(func() {
		expectedReturnList = []*grpcv1.Component{}
		newThing := corev1.PodList{Items: []corev1.Pod{}}
		allRunningPods = &newThing
	})

	Describe("no pods from current loadtest is running", func() {
		BeforeEach(func() {
			//Set up the expected returned list, in these cases should be full list
			for i := 0; i < len(currentLoadTest.Spec.Clients); i++ {
				expectedReturnList = append(expectedReturnList, &currentLoadTest.Spec.Clients[i].Component)
			}
			for i := 0; i < len(currentLoadTest.Spec.Servers); i++ {
				expectedReturnList = append(expectedReturnList, &currentLoadTest.Spec.Servers[i].Component)
			}
			expectedReturnList = append(expectedReturnList, &currentLoadTest.Spec.Driver.Component)
		})

		Context("allRunningPods is empty", func() {
			allRunningPods = &corev1.PodList{Items: []corev1.Pod{}}
			It("return the full list", func() {
				Expect(checkIfEqual(returnedList, expectedReturnList)).To(Equal(true))
			})
		})

		Context("irrelevant pods are running", func() {
			allRunningPods = &corev1.PodList{Items: []corev1.Pod{}}
			allRunningPods.Items = append(allRunningPods.Items, createPodListWithIrrelevantPod().Items...)
			It("return the full list", func() {
				Expect(checkIfEqual(returnedList, expectedReturnList)).To(Equal(true))
			})
		})
	})

	Describe("some of pods from current loadtest is running", func() {

		BeforeEach(func() {
			//Add pod with defaults.ComponentNameLabel: server-1, pod with defaults.ComponentNameLabel: client-2, driver
			//to running podlist
			allRunningPods.Items = append(allRunningPods.Items,
				corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "random_name",
						Labels: map[string]string{
							defaults.LoadTestLabel:      "test-loadtest-multiple-clients-and-servers",
							defaults.RoleLabel:          "server",
							defaults.ComponentNameLabel: "server-1",
						},
					},
				},
				corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "random_name",
						Labels: map[string]string{
							defaults.LoadTestLabel:      "test-loadtest-multiple-clients-and-servers",
							defaults.RoleLabel:          "client",
							defaults.ComponentNameLabel: "client-2",
						},
					},
				},
				corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "random_name",
						Labels: map[string]string{
							defaults.LoadTestLabel:      "test-loadtest-multiple-clients-and-servers",
							defaults.RoleLabel:          "driver",
							defaults.ComponentNameLabel: "driver-1",
						},
					},
				},
			)
			for i := 0; i < len(currentLoadTest.Spec.Clients); i++ {
				if *currentLoadTest.Spec.Clients[i].Name != "client-2" {
					expectedReturnList = append(expectedReturnList, &currentLoadTest.Spec.Clients[i].Component)
				}
			}

			for i := 0; i < len(currentLoadTest.Spec.Servers); i++ {
				if *currentLoadTest.Spec.Servers[i].Name != "server-1" {
					expectedReturnList = append(expectedReturnList, &currentLoadTest.Spec.Servers[i].Component)
				}
			}
		})

		Context("only pods from current loadtest are running", func() {
			It("return the a list of pods missing from collection of running pods", func() {
				Expect(checkIfEqual(returnedList, expectedReturnList)).To(Equal(true))
			})
		})

		Context("there are irrelevant pods running together", func() {
			allRunningPods.Items = append(allRunningPods.Items, createPodListWithIrrelevantPod().Items...)
			It("return the a list of pods missing from collection of running pods", func() {
				Expect(checkIfEqual(returnedList, expectedReturnList)).To(Equal(true))
			})
		})
	})

	Describe("all of pods from current loadtest is running", func() {

		BeforeEach(func() {
			allRunningPods = populatePodListWithCurrentLoadTestPod(currentLoadTest)
		})

		Context("only pods from current loadtest are running", func() {
			It("return empty list", func() {
				Expect(checkIfEqual(returnedList, expectedReturnList)).To(Equal(true))
			})
		})

		Context("there are irrelevant pods running together", func() {
			allRunningPods.Items = append(allRunningPods.Items, createPodListWithIrrelevantPod().Items...)
			It("should return empty list", func() {
				Expect(checkIfEqual(returnedList, expectedReturnList)).To(Equal(true))
			})
		})
	})

})

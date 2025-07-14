// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package slurmcontrol

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/puttsk/hostlist"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"k8s.io/utils/set"

	api "github.com/SlinkyProject/slurm-client/api/v0043"
	"github.com/SlinkyProject/slurm-client/pkg/client"
	"github.com/SlinkyProject/slurm-client/pkg/client/fake"
	"github.com/SlinkyProject/slurm-client/pkg/object"
	"github.com/SlinkyProject/slurm-client/pkg/types"

	slinkyv1alpha1 "github.com/SlinkyProject/slurm-operator/api/v1alpha1"
	nodesetutils "github.com/SlinkyProject/slurm-operator/internal/controller/nodeset/utils"
	"github.com/SlinkyProject/slurm-operator/internal/resources"
	"github.com/SlinkyProject/slurm-operator/internal/utils/podinfo"
)

func newNodeSet(name, clusterName string, replicas int32) *slinkyv1alpha1.NodeSet {
	return &slinkyv1alpha1.NodeSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: corev1.NamespaceDefault,
			Name:      name,
		},
		Spec: slinkyv1alpha1.NodeSetSpec{
			ClusterName: clusterName,
			Replicas:    &replicas,
		},
	}
}

func newSlurmClusters(clusterName string, client client.Client) *resources.Clusters {
	clusters := resources.NewClusters()
	key := k8stypes.NamespacedName{
		Namespace: corev1.NamespaceDefault,
		Name:      clusterName,
	}
	clusters.Add(key, client)
	return clusters
}

var _ = Describe("SlurmControlInterface", func() {
	const clusterName string = "slurm"
	var slurmcontrol SlurmControlInterface
	var nodeset *slinkyv1alpha1.NodeSet
	var pod *corev1.Pod
	var sclient client.Client

	updateFn := func(_ context.Context, obj object.Object, req any, opts ...client.UpdateOption) error {
		switch o := obj.(type) {
		case *types.V0043Node:
			r, ok := req.(api.V0043UpdateNodeMsg)
			if !ok {
				return errors.New("failed to cast request object")
			}
			stateSet := set.New(ptr.Deref(o.State, []api.V0043NodeState{})...)
			statesReq := ptr.Deref(r.State, []api.V0043UpdateNodeMsgState{})
			for _, stateReq := range statesReq {
				switch stateReq {
				case api.V0043UpdateNodeMsgStateUNDRAIN:
					stateSet.Delete(api.V0043NodeStateDRAIN)
				default:
					stateSet.Insert(api.V0043NodeState(stateReq))
				}
			}
			o.State = ptr.To(stateSet.UnsortedList())
			o.Comment = r.Comment
			o.Reason = r.Reason
		default:
			return errors.New("failed to cast slurm object")
		}
		return nil
	}

	Context("UpdateNodeWithPodInfo()", func() {
		It("Should update node comment with podInfo", func() {
			By("Setup initial system state")
			nodeset = newNodeSet("foo", clusterName, 1)
			pod = nodesetutils.NewNodeSetPod(nodeset, 0, "")
			slurmNodename := nodesetutils.GetNodeName(pod)
			node := &types.V0043Node{
				V0043Node: api.V0043Node{
					Name: ptr.To(slurmNodename),
					State: ptr.To([]api.V0043NodeState{
						api.V0043NodeStateIDLE,
					}),
				},
			}
			sclient = fake.NewClientBuilder().WithUpdateFn(updateFn).WithObjects(node).Build()
			clusters := newSlurmClusters(clusterName, sclient)
			slurmcontrol = NewSlurmControl(clusters)

			By("Update Slurm pod info")
			err := slurmcontrol.UpdateNodeWithPodInfo(ctx, nodeset, pod)
			Expect(err).ToNot(HaveOccurred())

			By("Check Slurm Node podInfo")
			wantPodInfo := podinfo.PodInfo{
				Namespace: pod.GetNamespace(),
				PodName:   pod.GetName(),
			}
			checkNode := &types.V0043Node{}
			key := object.ObjectKey(slurmNodename)
			err = sclient.Get(ctx, key, checkNode)
			Expect(err).ToNot(HaveOccurred())
			checkPodInfo := podinfo.PodInfo{}
			err = podinfo.ParseIntoPodInfo(checkNode.Comment, &checkPodInfo)
			Expect(err).ToNot(HaveOccurred())
			Expect(checkPodInfo.Equal(wantPodInfo)).To(BeTrue())
		})
	})

	Context("MakeNodeDrain()", func() {
		It("Should DRAIN the IDLE Slurm node", func() {
			By("Setup initial system state")
			nodeset = newNodeSet("foo", clusterName, 1)
			pod = nodesetutils.NewNodeSetPod(nodeset, 0, "")
			slurmNodename := nodesetutils.GetNodeName(pod)
			node := &types.V0043Node{
				V0043Node: api.V0043Node{
					Name: ptr.To(slurmNodename),
					State: ptr.To([]api.V0043NodeState{
						api.V0043NodeStateIDLE,
					}),
				},
			}
			sclient = fake.NewClientBuilder().WithUpdateFn(updateFn).WithObjects(node).Build()
			clusters := newSlurmClusters(clusterName, sclient)
			slurmcontrol = NewSlurmControl(clusters)

			By("Draining matching Slurm node")
			err := slurmcontrol.MakeNodeDrain(ctx, nodeset, pod, "drain")
			Expect(err).ToNot(HaveOccurred())

			By("Check Slurm Node state")
			checkNode := &types.V0043Node{}
			key := object.ObjectKey(slurmNodename)
			err = sclient.Get(ctx, key, checkNode)
			Expect(err).ToNot(HaveOccurred())
			isDrain := checkNode.GetStateAsSet().Has(api.V0043NodeStateDRAIN)
			Expect(isDrain).To(BeTrue())
		})
	})

	Context("MakeNodeUndrain()", func() {
		It("Should UNDRAIN the IDLE Slurm node", func() {
			By("Setup initial system state")
			nodeset = newNodeSet("foo", clusterName, 1)
			pod = nodesetutils.NewNodeSetPod(nodeset, 0, "")
			node := &types.V0043Node{
				V0043Node: api.V0043Node{
					Name: ptr.To(nodesetutils.GetNodeName(pod)),
					State: ptr.To([]api.V0043NodeState{
						api.V0043NodeStateIDLE,
						api.V0043NodeStateDRAIN,
					}),
				},
			}
			sclient = fake.NewClientBuilder().WithUpdateFn(updateFn).WithObjects(node).Build()
			clusters := newSlurmClusters(clusterName, sclient)
			slurmcontrol = NewSlurmControl(clusters)

			By("Draining matching Slurm node")
			err := slurmcontrol.MakeNodeUndrain(ctx, nodeset, pod, "undrain")
			Expect(err).ToNot(HaveOccurred())

			By("Check Slurm Node state")
			checkNode := &types.V0043Node{}
			key := object.ObjectKey(nodesetutils.GetNodeName(pod))
			err = sclient.Get(ctx, key, checkNode)
			Expect(err).ToNot(HaveOccurred())
			isundrain := !checkNode.GetStateAsSet().Has(api.V0043NodeStateDRAIN)
			Expect(isundrain).To(BeTrue())
		})
	})

	Context("GetNodeDeadlines()", func() {
		now := time.Now()

		It("Should get completion time for jobs", func() {
			By("Setup initial system state")
			nodeset = newNodeSet("bar", clusterName, 1)
			pod = nodesetutils.NewNodeSetPod(nodeset, 0, "")
			pod2 := nodesetutils.NewNodeSetPod(nodeset, 1, "")
			pods := []*corev1.Pod{pod, pod2}
			nodeList := &types.V0043NodeList{
				Items: []types.V0043Node{
					{
						V0043Node: api.V0043Node{
							Name: ptr.To(nodesetutils.GetNodeName(pod)),
							State: ptr.To([]api.V0043NodeState{
								api.V0043NodeStateMIXED,
							}),
						},
					},
					{
						V0043Node: api.V0043Node{
							Name: ptr.To(nodesetutils.GetNodeName(pod2)),
							State: ptr.To([]api.V0043NodeState{
								api.V0043NodeStateMIXED,
							}),
						},
					},
				},
			}
			jobList := &types.V0043JobInfoList{
				Items: []types.V0043JobInfo{
					{
						V0043JobInfo: api.V0043JobInfo{
							JobId:     ptr.To[int32](1),
							JobState:  ptr.To([]api.V0043JobInfoJobState{api.V0043JobInfoJobStateRUNNING}),
							StartTime: ptr.To(api.V0043Uint64NoValStruct{Number: ptr.To(now.Unix())}),
							TimeLimit: ptr.To(api.V0043Uint32NoValStruct{Number: ptr.To(30 * int32(time.Minute.Seconds()))}),
							Nodes: func() *string {
								hostlist, err := hostlist.Compress([]string{*nodeList.Items[0].Name})
								if err != nil {
									panic(err)
								}
								return ptr.To(hostlist)
							}(),
						},
					},
					{
						V0043JobInfo: api.V0043JobInfo{
							JobId:     ptr.To[int32](2),
							JobState:  ptr.To([]api.V0043JobInfoJobState{api.V0043JobInfoJobStateRUNNING}),
							StartTime: ptr.To(api.V0043Uint64NoValStruct{Number: ptr.To(now.Unix())}),
							TimeLimit: ptr.To(api.V0043Uint32NoValStruct{Number: ptr.To(45 * int32(time.Minute.Seconds()))}),
							Nodes: func() *string {
								hostlist, err := hostlist.Compress([]string{*nodeList.Items[0].Name, *nodeList.Items[1].Name})
								if err != nil {
									panic(err)
								}
								return ptr.To(hostlist)
							}(),
						},
					},
					{
						V0043JobInfo: api.V0043JobInfo{
							JobId:     ptr.To[int32](3),
							JobState:  ptr.To([]api.V0043JobInfoJobState{api.V0043JobInfoJobStateRUNNING}),
							StartTime: ptr.To(api.V0043Uint64NoValStruct{Number: ptr.To(now.Unix())}),
							TimeLimit: ptr.To(api.V0043Uint32NoValStruct{Number: ptr.To(int32(time.Hour.Seconds()))}),
							Nodes: func() *string {
								hostlist, err := hostlist.Compress([]string{*nodeList.Items[0].Name})
								if err != nil {
									panic(err)
								}
								return ptr.To(hostlist)
							}(),
						},
					},
					{
						V0043JobInfo: api.V0043JobInfo{
							JobId:    ptr.To[int32](4),
							JobState: ptr.To([]api.V0043JobInfoJobState{api.V0043JobInfoJobStateCOMPLETED}),
							Nodes: func() *string {
								hostlist, err := hostlist.Compress([]string{*nodeList.Items[0].Name, *nodeList.Items[1].Name})
								if err != nil {
									panic(err)
								}
								return ptr.To(hostlist)
							}(),
						},
					},
					{
						V0043JobInfo: api.V0043JobInfo{
							JobId:    ptr.To[int32](5),
							JobState: ptr.To([]api.V0043JobInfoJobState{api.V0043JobInfoJobStateCOMPLETED}),
							Nodes: func() *string {
								hostlist, err := hostlist.Compress([]string{*nodeList.Items[1].Name})
								if err != nil {
									panic(err)
								}
								return ptr.To(hostlist)
							}(),
						},
					},
				},
			}
			sclient = fake.NewClientBuilder().WithLists(nodeList, jobList).Build()
			clusters := newSlurmClusters(clusterName, sclient)
			slurmcontrol = NewSlurmControl(clusters)

			By("Getting TimeStore")
			ts, err := slurmcontrol.GetNodeDeadlines(ctx, nodeset, pods)
			Expect(err).ToNot(HaveOccurred())

			By("Check TimeStore for Slurm Nodes")
			for _, node := range nodeList.Items {
				Expect(ts.Peek(*node.Name).After(now)).To(BeTrue())
			}
		})
	})
})

func Test_realSlurmControl_IsNodeDrain(t *testing.T) {
	ctx := context.Background()
	const clusterName string = "slurm"
	nodeset := newNodeSet("foo", clusterName, 1)
	pod := nodesetutils.NewNodeSetPod(nodeset, 0, "")
	type fields struct {
		slurmClusters *resources.Clusters
	}
	type args struct {
		ctx     context.Context
		nodeset *slinkyv1alpha1.NodeSet
		pod     *corev1.Pod
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Not DRAIN",
			fields: func() fields {
				node := &types.V0043Node{
					V0043Node: api.V0043Node{
						Name: ptr.To(nodesetutils.GetNodeName(pod)),
						State: ptr.To([]api.V0043NodeState{
							api.V0043NodeStateIDLE,
						}),
					},
				}
				sclient := fake.NewClientBuilder().WithObjects(node).Build()
				return fields{
					slurmClusters: newSlurmClusters(clusterName, sclient),
				}
			}(),
			args: args{
				ctx:     ctx,
				nodeset: nodeset,
				pod:     pod,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Is DRAIN",
			fields: func() fields {
				node := &types.V0043Node{
					V0043Node: api.V0043Node{
						Name: ptr.To(nodesetutils.GetNodeName(pod)),
						State: ptr.To([]api.V0043NodeState{
							api.V0043NodeStateDRAIN,
						}),
					},
				}
				sclient := fake.NewClientBuilder().WithObjects(node).Build()
				return fields{
					slurmClusters: newSlurmClusters(clusterName, sclient),
				}
			}(),
			args: args{
				ctx:     ctx,
				nodeset: nodeset,
				pod:     pod,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &realSlurmControl{
				slurmClusters: tt.fields.slurmClusters,
			}
			got, err := r.IsNodeDrain(tt.args.ctx, tt.args.nodeset, tt.args.pod)
			if (err != nil) != tt.wantErr {
				t.Errorf("realSlurmControl.IsNodeDrain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("realSlurmControl.IsNodeDrain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_realSlurmControl_IsNodeDrained(t *testing.T) {
	ctx := context.Background()
	const clusterName string = "slurm"
	nodeset := newNodeSet("foo", clusterName, 1)
	pod := nodesetutils.NewNodeSetPod(nodeset, 0, "")
	type fields struct {
		slurmClusters *resources.Clusters
	}
	type args struct {
		ctx     context.Context
		nodeset *slinkyv1alpha1.NodeSet
		pod     *corev1.Pod
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "IDLE",
			fields: func() fields {
				node := &types.V0043Node{
					V0043Node: api.V0043Node{
						Name: ptr.To(nodesetutils.GetNodeName(pod)),
						State: ptr.To([]api.V0043NodeState{
							api.V0043NodeStateIDLE,
						}),
					},
				}
				sclient := fake.NewClientBuilder().WithObjects(node).Build()
				return fields{
					slurmClusters: newSlurmClusters(clusterName, sclient),
				}
			}(),
			args: args{
				ctx:     ctx,
				nodeset: nodeset,
				pod:     pod,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "IDLE+DRAIN",
			fields: func() fields {
				node := &types.V0043Node{
					V0043Node: api.V0043Node{
						Name: ptr.To(nodesetutils.GetNodeName(pod)),
						State: ptr.To([]api.V0043NodeState{
							api.V0043NodeStateIDLE,
							api.V0043NodeStateDRAIN,
						}),
					},
				}
				sclient := fake.NewClientBuilder().WithObjects(node).Build()
				return fields{
					slurmClusters: newSlurmClusters(clusterName, sclient),
				}
			}(),
			args: args{
				ctx:     ctx,
				nodeset: nodeset,
				pod:     pod,
			},
			want: true,
		},
		{
			name: "ALLOC+DRAIN",
			fields: func() fields {
				node := &types.V0043Node{
					V0043Node: api.V0043Node{
						Name: ptr.To(nodesetutils.GetNodeName(pod)),
						State: ptr.To([]api.V0043NodeState{
							api.V0043NodeStateALLOCATED,
							api.V0043NodeStateDRAIN,
						}),
					},
				}
				sclient := fake.NewClientBuilder().WithObjects(node).Build()
				return fields{
					slurmClusters: newSlurmClusters(clusterName, sclient),
				}
			}(),
			args: args{
				ctx:     ctx,
				nodeset: nodeset,
				pod:     pod,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "DOWN+DRAIN",
			fields: func() fields {
				node := &types.V0043Node{
					V0043Node: api.V0043Node{
						Name: ptr.To(nodesetutils.GetNodeName(pod)),
						State: ptr.To([]api.V0043NodeState{
							api.V0043NodeStateDOWN,
							api.V0043NodeStateDRAIN,
						}),
					},
				}
				sclient := fake.NewClientBuilder().WithObjects(node).Build()
				return fields{
					slurmClusters: newSlurmClusters(clusterName, sclient),
				}
			}(),
			args: args{
				ctx:     ctx,
				nodeset: nodeset,
				pod:     pod,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &realSlurmControl{
				slurmClusters: tt.fields.slurmClusters,
			}
			got, err := r.IsNodeDrained(tt.args.ctx, tt.args.nodeset, tt.args.pod)
			if (err != nil) != tt.wantErr {
				t.Errorf("realSlurmControl.IsNodeDrained() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("realSlurmControl.IsNodeDrained() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_realSlurmControl_CalculateNodeStatus(t *testing.T) {
	ctx := context.Background()
	const clusterName string = "slurm"
	nodeset := newNodeSet("foo", clusterName, 1)
	nodeset2 := newNodeSet("baz", clusterName, 1)
	type fields struct {
		slurmClusters *resources.Clusters
	}
	type args struct {
		ctx     context.Context
		nodeset *slinkyv1alpha1.NodeSet
		pods    []*corev1.Pod
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    SlurmNodeStatus
		wantErr bool
	}{
		{
			name: "Empty",
			fields: func() fields {
				nodeList := &types.V0043NodeList{
					Items: []types.V0043Node{},
				}
				sclient := fake.NewClientBuilder().WithLists(nodeList).Build()
				return fields{
					slurmClusters: newSlurmClusters(clusterName, sclient),
				}
			}(),
			args: args{
				ctx:     ctx,
				nodeset: nodeset,
				pods:    []*corev1.Pod{},
			},
			want:    SlurmNodeStatus{},
			wantErr: false,
		},
		{
			name: "Different NodeSets",
			fields: func() fields {
				nodeList := &types.V0043NodeList{
					Items: []types.V0043Node{
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 0, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateIDLE,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset2, 0, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateIDLE,
								}),
							},
						},
					},
				}
				sclient := fake.NewClientBuilder().WithLists(nodeList).Build()
				return fields{
					slurmClusters: newSlurmClusters(clusterName, sclient),
				}
			}(),
			args: args{
				ctx:     ctx,
				nodeset: nodeset,
				pods: []*corev1.Pod{
					nodesetutils.NewNodeSetPod(nodeset, 0, ""),
				},
			},
			want: SlurmNodeStatus{
				Total: 1,

				Idle: 1,
			},
			wantErr: false,
		},
		{
			name: "Only base state",
			fields: func() fields {
				nodeList := &types.V0043NodeList{
					Items: []types.V0043Node{
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 0, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateIDLE,
								}),
							},
						},
					},
				}
				sclient := fake.NewClientBuilder().WithLists(nodeList).Build()
				return fields{
					slurmClusters: newSlurmClusters(clusterName, sclient),
				}
			}(),
			args: args{
				ctx:     ctx,
				nodeset: nodeset,
				pods: []*corev1.Pod{
					nodesetutils.NewNodeSetPod(nodeset, 0, ""),
				},
			},
			want: SlurmNodeStatus{
				Total: 1,

				Idle: 1,
			},
			wantErr: false,
		},
		{
			name: "Base and flag state",
			fields: func() fields {
				nodeList := &types.V0043NodeList{
					Items: []types.V0043Node{
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 0, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateIDLE,
									api.V0043NodeStateDRAIN,
								}),
							},
						},
					},
				}
				sclient := fake.NewClientBuilder().WithLists(nodeList).Build()
				return fields{
					slurmClusters: newSlurmClusters(clusterName, sclient),
				}
			}(),
			args: args{
				ctx:     ctx,
				nodeset: nodeset,
				pods: []*corev1.Pod{
					nodesetutils.NewNodeSetPod(nodeset, 0, ""),
				},
			},
			want: SlurmNodeStatus{
				Total: 1,

				Idle:  1,
				Drain: 1,
			},
			wantErr: false,
		},
		{
			name: "All base states",
			fields: func() fields {
				nodeList := &types.V0043NodeList{
					Items: []types.V0043Node{
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 0, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateALLOCATED,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 1, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateDOWN,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 2, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateERROR,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 3, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateFUTURE,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 4, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateIDLE,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 5, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateMIXED,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 6, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateUNKNOWN,
								}),
							},
						},
					},
				}
				sclient := fake.NewClientBuilder().WithLists(nodeList).Build()
				return fields{
					slurmClusters: newSlurmClusters(clusterName, sclient),
				}
			}(),
			args: args{
				ctx:     ctx,
				nodeset: nodeset,
				pods: []*corev1.Pod{
					nodesetutils.NewNodeSetPod(nodeset, 0, ""),
					nodesetutils.NewNodeSetPod(nodeset, 1, ""),
					nodesetutils.NewNodeSetPod(nodeset, 2, ""),
					nodesetutils.NewNodeSetPod(nodeset, 3, ""),
					nodesetutils.NewNodeSetPod(nodeset, 4, ""),
					nodesetutils.NewNodeSetPod(nodeset, 5, ""),
					nodesetutils.NewNodeSetPod(nodeset, 6, ""),
				},
			},
			want: SlurmNodeStatus{
				Total: 7,

				Allocated: 1,
				Down:      1,
				Error:     1,
				Future:    1,
				Idle:      1,
				Mixed:     1,
				Unknown:   1,
			},
			wantErr: false,
		},
		{
			name: "All flag states",
			fields: func() fields {
				nodeList := &types.V0043NodeList{
					Items: []types.V0043Node{
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 0, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateCOMPLETING,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 1, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateDRAIN,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 2, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateFAIL,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 3, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateINVALID,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 4, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateINVALIDREG,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 5, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateMAINTENANCE,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 6, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateNOTRESPONDING,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 7, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateUNDRAIN,
								}),
							},
						},
					},
				}
				sclient := fake.NewClientBuilder().WithLists(nodeList).Build()
				return fields{
					slurmClusters: newSlurmClusters(clusterName, sclient),
				}
			}(),
			args: args{
				ctx:     ctx,
				nodeset: nodeset,
				pods: []*corev1.Pod{
					nodesetutils.NewNodeSetPod(nodeset, 0, ""),
					nodesetutils.NewNodeSetPod(nodeset, 1, ""),
					nodesetutils.NewNodeSetPod(nodeset, 2, ""),
					nodesetutils.NewNodeSetPod(nodeset, 3, ""),
					nodesetutils.NewNodeSetPod(nodeset, 4, ""),
					nodesetutils.NewNodeSetPod(nodeset, 5, ""),
					nodesetutils.NewNodeSetPod(nodeset, 6, ""),
					nodesetutils.NewNodeSetPod(nodeset, 7, ""),
				},
			},
			want: SlurmNodeStatus{
				Total: 8,

				Completing:    1,
				Drain:         1,
				Fail:          1,
				Invalid:       1,
				InvalidReg:    1,
				Maintenance:   1,
				NotResponding: 1,
				Undrain:       1,
			},
			wantErr: false,
		},
		{
			name: "All states",
			fields: func() fields {
				nodeList := &types.V0043NodeList{
					Items: []types.V0043Node{
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 0, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateALLOCATED,
									api.V0043NodeStateCOMPLETING,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 1, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateDOWN,
									api.V0043NodeStateDRAIN,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 2, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateERROR,
									api.V0043NodeStateFAIL,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 3, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateFUTURE,
									api.V0043NodeStateINVALID,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 4, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateFUTURE,
									api.V0043NodeStateINVALIDREG,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 5, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateIDLE,
									api.V0043NodeStateMAINTENANCE,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 6, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateMIXED,
									api.V0043NodeStateNOTRESPONDING,
								}),
							},
						},
						{
							V0043Node: api.V0043Node{
								Name: ptr.To(nodesetutils.GetNodeName(nodesetutils.NewNodeSetPod(nodeset, 7, ""))),
								State: ptr.To([]api.V0043NodeState{
									api.V0043NodeStateUNKNOWN,
									api.V0043NodeStateUNDRAIN,
								}),
							},
						},
					},
				}
				sclient := fake.NewClientBuilder().WithLists(nodeList).Build()
				return fields{
					slurmClusters: newSlurmClusters(clusterName, sclient),
				}
			}(),
			args: args{
				ctx:     ctx,
				nodeset: nodeset,
				pods: []*corev1.Pod{
					nodesetutils.NewNodeSetPod(nodeset, 0, ""),
					nodesetutils.NewNodeSetPod(nodeset, 1, ""),
					nodesetutils.NewNodeSetPod(nodeset, 2, ""),
					nodesetutils.NewNodeSetPod(nodeset, 3, ""),
					nodesetutils.NewNodeSetPod(nodeset, 4, ""),
					nodesetutils.NewNodeSetPod(nodeset, 5, ""),
					nodesetutils.NewNodeSetPod(nodeset, 6, ""),
					nodesetutils.NewNodeSetPod(nodeset, 7, ""),
				},
			},
			want: SlurmNodeStatus{
				Total: 8,

				Allocated: 1,
				Down:      1,
				Error:     1,
				Future:    2,
				Idle:      1,
				Mixed:     1,
				Unknown:   1,

				Completing:    1,
				Drain:         1,
				Fail:          1,
				Invalid:       1,
				InvalidReg:    1,
				Maintenance:   1,
				NotResponding: 1,
				Undrain:       1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &realSlurmControl{
				slurmClusters: tt.fields.slurmClusters,
			}
			got, err := r.CalculateNodeStatus(tt.args.ctx, tt.args.nodeset, tt.args.pods)
			if (err != nil) != tt.wantErr {
				t.Errorf("realSlurmControl.CalculateNodeStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !apiequality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("realSlurmControl.CalculateNodeStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tolerateError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Nil",
			args: args{
				err: nil,
			},
			want: true,
		},
		{
			name: "Empty",
			args: args{
				err: errors.New(""),
			},
			want: false,
		},
		{
			name: "NotFound",
			args: args{
				err: errors.New(http.StatusText(http.StatusNotFound)),
			},
			want: true,
		},
		{
			name: "NoContent",
			args: args{
				err: errors.New(http.StatusText(http.StatusNoContent)),
			},
			want: true,
		},
		{
			name: "Forbidden",
			args: args{
				err: errors.New(http.StatusText(http.StatusForbidden)),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tolerateError(tt.args.err); got != tt.want {
				t.Errorf("tolerateError() = %v, want %v", got, tt.want)
			}
		})
	}
}

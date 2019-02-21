/*
Copyright 2019 Cornelius Weig

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/api/authorization/v1"
	authv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"sync"
)

/*
restConfig, _ := flags.ToRESTConfig()
authClient, _ := authv1.NewForConfig(restConfig)
allowed, err := CheckResourceAccessPar(authClient, resources)
if err != nil {
	return nil, errors.Wrap(err, "check resource access")
}

allowedResourceTypes := ToResourceTypes(allowed)*/

func CheckResourceAccessPar(authClient *authv1.AuthorizationV1Client, grs []groupResource) (allowed []groupResource, err error) {
	reviews := authClient.SelfSubjectAccessReviews()

	allowedChan := make(chan groupResource)
	forbiddenChan := make(chan groupResource)

	group := sync.WaitGroup{}

	namespace := viper.GetString("namespace")
	for _, gr := range grs {
		namespace := namespace
		gr := gr
		group.Add(1)
		go func(allowed, forbidden chan<- groupResource) {
			defer group.Done()

			// This seems to be a bug in kubernetes. If namespace is set for non-namespaced
			// resources, the access is reported as "allowed", but in fact it is forbidden.
			if !gr.APIResource.Namespaced {
				namespace = ""
			}

			review := v1.SelfSubjectAccessReview{
				Spec: v1.SelfSubjectAccessReviewSpec{
					ResourceAttributes: &v1.ResourceAttributes{
						Verb:      "list",
						Resource:  gr.APIResource.Name,
						Group:     gr.APIGroup,
						Namespace: namespace,
					},
				},
			}
			accessReview, e := reviews.Create(&review)
			if e != nil {
				err = errors.Wrap(e, "retrieve authority")
				return
			}
			logrus.Info(accessReview)
			if accessReview.Status.Allowed {
				allowed <- gr
			} else {
				forbidden <- gr
			}
		}(allowedChan, forbiddenChan)
	}

	forbidden := []groupResource{}
	go func(c <-chan groupResource) {
		for gr := range c {
			forbidden = append(forbidden, gr)
		}
	}(forbiddenChan)
	go func(c <-chan groupResource) {
		for gr := range c {
			allowed = append(allowed, gr)
		}
	}(allowedChan)

	group.Wait()

	close(allowedChan)
	close(forbiddenChan)

	if len(forbidden) > 0 {
		logrus.Warnf("The following resources may not be read: %s", ToResourceTypes(forbidden))
	}

	logrus.Debugf("Readable: %s", ToResourceTypes(allowed))

	return
}

func CheckResourceAccess(authClient *authv1.AuthorizationV1Client, grs []groupResource) (allowed []groupResource, err error) {
	forbidden := []groupResource{}
	reviews := authClient.SelfSubjectAccessReviews()

	namespace := viper.GetString("namespace")
	for _, gr := range grs {
		namespace := namespace
		gr := gr
		// This seems to be a bug in kubernetes. If namespace is set for non-namespaced
		// resources, the access is reported as "allowed", but in fact it is forbidden.
		if !gr.APIResource.Namespaced {
			namespace = ""
		}

		review := v1.SelfSubjectAccessReview{
			Spec: v1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &v1.ResourceAttributes{
					Verb:      "list",
					Resource:  gr.APIResource.Name,
					Group:     gr.APIGroup,
					Namespace: namespace,
				},
			},
		}
		accessReview, e := reviews.Create(&review)
		if e != nil {
			err = errors.Wrap(e, "retrieve authority")
			return
		}
		logrus.Info(accessReview)
		if accessReview.Status.Allowed {
			allowed = append(allowed, gr)
		} else {
			forbidden = append(forbidden, gr)
		}
	}

	if len(forbidden) > 0 {
		logrus.Warnf("The following resources may not be read: %s", ToResourceTypes(forbidden))
	}

	logrus.Debugf("Readable: %s", ToResourceTypes(allowed))

	return
}

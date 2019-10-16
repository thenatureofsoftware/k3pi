package cmd

import (
	"github.com/TheNatureOfSoftware/k3pi/mocks"
	"github.com/TheNatureOfSoftware/k3pi/pkg/client"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/TheNatureOfSoftware/k3pi/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestUpgradeK3s(t *testing.T) {
	cf, _ := client.NewFakeClientFactory()
	err := upgradeK3s(cf, "v0.9.1", test.CreateNodes()[0:1], false)
	if err != nil {
		t.Error(err)
	}
}

func TestUpgrade_UpgradeK3OS(t *testing.T) {

	taskType := "*install.OSUpgradeTask"
	_, mockInstallerFactory, mockInstaller := setUpMocks(taskType)

	cf, _ := client.NewFakeClientFactory()
	err := Upgrade(cf, COS, "v0.4.0", test.CreateNodes(), false)

	assert.Nil(t, err, "unexpected error")
	assertMocks(mockInstallerFactory, t, mockInstaller)
}

func TestUpgrade_UpgradeK3s(t *testing.T) {

	taskType := "*install.K3sUpgradeTask"
	_, mockInstallerFactory, mockInstaller := setUpMocks(taskType)

	cf, _ := client.NewFakeClientFactory()
	err := Upgrade(cf, CK3s, "v0.9.1", test.CreateNodes(), false)

	assert.Nil(t, err, "unexpected error")
	assertMocks(mockInstallerFactory, t, mockInstaller)
}

func assertMocks(mockInstallerFactory *mocks.InstallerFactory, t *testing.T, mockInstaller *mocks.Installer) {
	mockInstallerFactory.AssertExpectations(t)
	mockInstallerFactory.AssertExpectations(t)
	mockInstaller.AssertExpectations(t)
}

func setUpMocks(taskType string) (*mocks.InstallerFactories, *mocks.InstallerFactory, *mocks.Installer) {
	mockInstallerFactories := new(mocks.InstallerFactories)
	mockInstallerFactory := new(mocks.InstallerFactory)
	mockInstaller := new(mocks.Installer)
	installerFactories = mockInstallerFactories

	mockInstallerFactories.On("GetFactory", mock.AnythingOfType(taskType)).Return(mockInstallerFactory).Times(1)
	mockInstallerFactory.On("MakeInstallers", mock.AnythingOfType(taskType), mock.AnythingOfType("string")).Return(model.Installers{mockInstaller}).Times(1)
	mockInstaller.On("Install").Return(nil)

	return mockInstallerFactories, mockInstallerFactory, mockInstaller
}

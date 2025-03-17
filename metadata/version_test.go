package metadata

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 定义更丰富的测试数据
var enhancedTestVersionData = MetadataMap{
	"test-module": Metadata{
		VersionMappings: []*VersionMapping{
			{CVersion: "1.7.18", GoVersions: []string{"v1.2.0", "v1.2.1"}},
			{CVersion: "1.7.19", GoVersions: []string{"v1.3.0"}},
			{CVersion: "1.8.0", GoVersions: []string{"v1.4.0", "v1.4.1"}},
		},
	},
	"empty-module": Metadata{VersionMappings: []*VersionMapping{}},
}

// 设置测试环境
func setupTestEnv(t *testing.T, testData MetadataMap) (*metadataMgr, func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(testData)
	}))

	originalURL := remoteMetadataURL
	remoteMetadataURL = server.URL

	tmpDir := t.TempDir()
	mgr, err := NewMetadataMgr(tmpDir)
	require.NoError(t, err)

	cleanup := func() {
		server.Close()
		remoteMetadataURL = originalURL
	}

	return mgr, cleanup
}

// TestLatestCVer 测试 LatestCVer 函数的实现
func TestLatestCVer(t *testing.T) {
	mgr, cleanup := setupTestEnv(t, enhancedTestVersionData)
	defer cleanup()

	// 测试正常情况
	latestCVer, err := mgr.LatestCVer("test-module")
	require.NoError(t, err)
	assert.Equal(t, "1.8.0", latestCVer)

	// 测试空模块
	_, err = mgr.LatestCVer("empty-module")
	assert.Error(t, err)

	// 测试不存在的模块
	_, err = mgr.LatestCVer("non-existent-module")
	assert.Error(t, err)
}

// TestLatestGoVer 测试 LatestGoVer 函数的实现
func TestLatestGoVer(t *testing.T) {
	mgr, cleanup := setupTestEnv(t, enhancedTestVersionData)
	defer cleanup()

	// 测试正常情况
	latestGoVer, err := mgr.LatestGoVer("test-module")
	require.NoError(t, err)
	assert.Equal(t, "v1.4.1", latestGoVer)

	// 测试空模块
	_, err = mgr.LatestGoVer("empty-module")
	assert.Error(t, err)

	// 测试不存在的模块
	_, err = mgr.LatestGoVer("non-existent-module")
	assert.Error(t, err)
}

// TestLatestGoVerFromCVer 测试 LatestGoVerFromCVer 函数的实现
func TestLatestGoVerFromCVer(t *testing.T) {
	mgr, cleanup := setupTestEnv(t, enhancedTestVersionData)
	defer cleanup()

	// 测试正常情况
	latestGoVer, err := mgr.LatestGoVerFromCVer("test-module", "1.7.18")
	require.NoError(t, err)
	assert.Equal(t, "v1.2.1", latestGoVer)

	// 测试不存在的 C 版本
	_, err = mgr.LatestGoVerFromCVer("test-module", "non-existent-version")
	assert.Error(t, err)

	// 测试不存在的模块
	_, err = mgr.LatestGoVerFromCVer("non-existent-module", "1.7.18")
	assert.Error(t, err)
}

// TestGoVersFromCVer 测试 GoVersFromCVer 函数的实现
func TestGoVersFromCVer(t *testing.T) {
	mgr, cleanup := setupTestEnv(t, enhancedTestVersionData)
	defer cleanup()

	// 测试正常情况
	goVers, err := mgr.GoVersFromCVer("test-module", "1.7.18")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"v1.2.0", "v1.2.1"}, goVers)

	// 测试不存在的 C 版本
	_, err = mgr.GoVersFromCVer("test-module", "non-existent-version")
	assert.Error(t, err)

	// 测试不存在的模块
	_, err = mgr.GoVersFromCVer("non-existent-module", "1.7.18")
	assert.Error(t, err)
}

// TestCVerFromGoVer 测试 CVerFromGoVer 函数的实现
func TestCVerFromGoVer(t *testing.T) {
	mgr, cleanup := setupTestEnv(t, enhancedTestVersionData)
	defer cleanup()

	// 测试正常情况
	cVer, err := mgr.CVerFromGoVer("test-module", "v1.3.0")
	require.NoError(t, err)
	assert.Equal(t, "1.7.19", cVer)

	// 测试不存在的 Go 版本
	_, err = mgr.CVerFromGoVer("test-module", "non-existent-version")
	assert.Error(t, err)

	// 测试不存在的模块
	_, err = mgr.CVerFromGoVer("non-existent-module", "v1.3.0")
	assert.Error(t, err)
}

// TestAllGoVersFromName 测试 AllGoVersFromName 函数的实现
func TestAllGoVersFromName(t *testing.T) {
	mgr, cleanup := setupTestEnv(t, enhancedTestVersionData)
	defer cleanup()

	// 测试正常情况
	goVers, err := mgr.AllGoVersFromName("test-module")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"v1.2.0", "v1.2.1", "v1.3.0", "v1.4.0", "v1.4.1"}, goVers)

	// 测试空模块
	goVers, err = mgr.AllGoVersFromName("empty-module")
	require.NoError(t, err)
	assert.Empty(t, goVers)

	// 测试不存在的模块
	_, err = mgr.AllGoVersFromName("non-existent-module")
	assert.Error(t, err)
}

// TestAllCVersFromName 测试 AllCVersFromName 函数的实现
func TestAllCVersFromName(t *testing.T) {
	mgr, cleanup := setupTestEnv(t, enhancedTestVersionData)
	defer cleanup()

	// 测试正常情况
	cVers, err := mgr.AllCVersFromName("test-module")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"1.7.18", "1.7.19", "1.8.0"}, cVers)

	// 测试空模块
	cVers, err = mgr.AllCVersFromName("empty-module")
	require.NoError(t, err)
	assert.Empty(t, cVers)

	// 测试不存在的模块
	_, err = mgr.AllCVersFromName("non-existent-module")
	assert.Error(t, err)
}

// TestAllVersionMappingsFromName 测试 AllVersionMappingsFromName 函数的实现
func TestAllVersionMappingsFromName(t *testing.T) {
	mgr, cleanup := setupTestEnv(t, enhancedTestVersionData)
	defer cleanup()

	// 测试正常情况
	mappings, err := mgr.AllVersionMappingsFromName("test-module")
	require.NoError(t, err)
	assert.Len(t, mappings, 3)

	// 测试空模块
	mappings, err = mgr.AllVersionMappingsFromName("empty-module")
	require.NoError(t, err)
	assert.Empty(t, mappings)

	// 测试不存在的模块
	_, err = mgr.AllVersionMappingsFromName("non-existent-module")
	assert.Error(t, err)
}

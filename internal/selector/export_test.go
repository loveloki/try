package selector

import "testing"

// SetupTestDirsForTest 供 selector_test 外部测试包搭建临时目录。
func SetupTestDirsForTest(t *testing.T) string {
	t.Helper()
	return setupTestDirs(t)
}

// DriveWithDialogFactory 驱动选择器并注入对话框工厂（供跨包集成测试使用）。
func DriveWithDialogFactory(t *testing.T, cfg Config, factory DialogFactory) SelectorModel {
	t.Helper()
	return driveModelWithFactory(t, cfg, factory)
}

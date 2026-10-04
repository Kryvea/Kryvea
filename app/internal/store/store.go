package store

import "context"

type Store interface {
	Assessment() AssessmentStore
	Category() CategoryStore
	Customer() CustomerStore
	FileReference() FileReferenceStore
	Poc() PocStore
	Setting() SettingStore
	Target() TargetStore
	Template() TemplateStore
	User() UserStore
	Vulnerability() VulnerabilityStore

	RunInTx(ctx context.Context, fn func(context.Context) (any, error)) (any, error)
	RunInTxWithLock(ctx context.Context, lockName string, fn func(context.Context) (any, error)) (any, error)
}

package acltree

import (
	"context"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/pkg/acl/account"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/pkg/acl/aclchanges/aclpb"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/pkg/acl/treestorage"
	"sync"
)

type AddResultSummary int

const (
	AddResultSummaryNothing AddResultSummary = iota
	AddResultSummaryAppend
	AddResultSummaryRebuild
)

type AddResult struct {
	OldHeads []string
	Heads    []string
	// TODO: add summary for changes
	Summary AddResultSummary
}

type TreeUpdateListener interface {
	Update(tree ACLTree)
	Rebuild(tree ACLTree)
}

type NoOpListener struct{}

func (n NoOpListener) Update(tree ACLTree) {}

func (n NoOpListener) Rebuild(tree ACLTree) {}

type ACLTree interface {
	ACLState() *ACLState
	AddContent(ctx context.Context, f func(builder ChangeBuilder) error) (*Change, error)
	AddChanges(ctx context.Context, changes ...*Change) (AddResult, error)
	AddRawChanges(ctx context.Context, changes ...*aclpb.RawChange) (AddResult, error)
	Heads() []string
	Root() *Change
	Iterate(func(change *Change) bool)
	IterateFrom(string, func(change *Change) bool)
	HasChange(string) bool

	Close() error
}

type aclTree struct {
	treeStorage    treestorage.TreeStorage
	accountData    *account.AccountData
	updateListener TreeUpdateListener

	fullTree         *Tree
	aclTreeFromStart *Tree // TODO: right now we don't use it, we can probably have only local var for now. This tree is built from start of the document
	aclState         *ACLState

	treeBuilder       *treeBuilder
	aclTreeBuilder    *aclTreeBuilder
	aclStateBuilder   *aclStateBuilder
	snapshotValidator *snapshotValidator
	changeBuilder     *changeBuilder

	sync.RWMutex
}

func BuildACLTree(
	t treestorage.TreeStorage,
	acc *account.AccountData,
	listener TreeUpdateListener) (ACLTree, error) {
	aclTreeBuilder := newACLTreeBuilder(t, acc.Decoder)
	treeBuilder := newTreeBuilder(t, acc.Decoder)
	snapshotValidator := newSnapshotValidator(acc.Decoder, acc) // TODO: this looks weird, change it
	aclStateBuilder := newACLStateBuilder(acc.Decoder, acc)
	changeBuilder := newChangeBuilder()

	aclTree := &aclTree{
		treeStorage:       t,
		accountData:       acc,
		fullTree:          nil,
		aclState:          nil,
		treeBuilder:       treeBuilder,
		aclTreeBuilder:    aclTreeBuilder,
		aclStateBuilder:   aclStateBuilder,
		snapshotValidator: snapshotValidator,
		changeBuilder:     changeBuilder,
		updateListener:    listener,
	}
	err := aclTree.rebuildFromStorage(false)
	if err != nil {
		return nil, err
	}
	err = aclTree.removeOrphans()
	if err != nil {
		return nil, err
	}
	err = t.SetHeads(aclTree.Heads())
	if err != nil {
		return nil, err
	}

	listener.Rebuild(aclTree)

	return aclTree, nil
}

// TODO: this is not used for now, in future we should think about not making full tree rebuild
//func (a *aclTree) rebuildFromTree(validateSnapshot bool) (err error) {
//	if validateSnapshot {
//		err = a.snapshotValidator.Init(a.aclTreeFromStart)
//		if err != nil {
//			return err
//		}
//
//		valid, err := a.snapshotValidator.ValidateSnapshot(a.fullTree.root)
//		if err != nil {
//			return err
//		}
//		if !valid {
//			return a.rebuildFromStorage(true)
//		}
//	}
//
//	err = a.aclStateBuilder.Init(a.fullTree)
//	if err != nil {
//		return err
//	}
//
//	a.aclState, err = a.aclStateBuilder.Build()
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

func (a *aclTree) removeOrphans() error {
	// removing attached or invalid orphans
	var toRemove []string

	orphans, err := a.treeStorage.Orphans()
	if err != nil {
		return err
	}
	for _, orphan := range orphans {
		if _, exists := a.fullTree.attached[orphan]; exists {
			toRemove = append(toRemove, orphan)
		}
		if _, exists := a.fullTree.invalidChanges[orphan]; exists {
			toRemove = append(toRemove, orphan)
		}
	}
	return a.treeStorage.RemoveOrphans(toRemove...)
}

func (a *aclTree) rebuildFromStorage(fromStart bool) error {
	a.treeBuilder.Init()
	a.aclTreeBuilder.Init()

	var err error
	a.fullTree, err = a.treeBuilder.Build(fromStart)
	if err != nil {
		return err
	}

	// TODO: remove this from context as this is used only to validate snapshot
	a.aclTreeFromStart, err = a.aclTreeBuilder.Build()
	if err != nil {
		return err
	}

	if !fromStart {
		err = a.snapshotValidator.Init(a.aclTreeFromStart)
		if err != nil {
			return err
		}

		valid, err := a.snapshotValidator.ValidateSnapshot(a.fullTree.root)
		if err != nil {
			return err
		}
		if !valid {
			return a.rebuildFromStorage(true)
		}
	}
	// TODO: there is a question how we can validate not only that the full tree is built correctly
	//  but also that the ACL prev ids are not messed up. I think we should probably compare the resulting
	//  acl state with the acl state which is built in aclTreeFromStart

	err = a.aclStateBuilder.Init(a.fullTree)
	if err != nil {
		return err
	}

	a.aclState, err = a.aclStateBuilder.Build()
	if err != nil {
		return err
	}

	return nil
}

func (a *aclTree) ACLState() *ACLState {
	// TODO: probably locks should be happening outside because we are using object cache
	a.RLock()
	defer a.RUnlock()
	return a.aclState
}

func (a *aclTree) AddContent(ctx context.Context, build func(builder ChangeBuilder) error) (*Change, error) {
	// TODO: add snapshot creation logic
	a.Lock()
	defer func() {
		a.Unlock()
		// TODO: should this be called in a separate goroutine to prevent accidental cycles (tree->updater->tree)
		a.updateListener.Update(a)
	}()

	a.changeBuilder.Init(a.aclState, a.fullTree, a.accountData)
	err := build(a.changeBuilder)
	if err != nil {
		return nil, err
	}

	ch, marshalled, err := a.changeBuilder.BuildAndApply()
	if err != nil {
		return nil, err
	}
	a.fullTree.AddFast(ch)

	err = a.treeStorage.AddRawChange(&aclpb.RawChange{
		Payload:   marshalled,
		Signature: ch.Signature(),
		Id:        ch.Id,
	})
	if err != nil {
		return nil, err
	}

	err = a.treeStorage.SetHeads([]string{ch.Id})
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (a *aclTree) AddRawChanges(ctx context.Context, rawChanges ...*aclpb.RawChange) (AddResult, error) {
	var aclChanges []*Change
	for _, ch := range rawChanges {
		change, err := NewFromRawChange(ch)
		// TODO: think what if we will have incorrect signatures on rawChanges, how everything will work
		if err != nil {
			continue
		}
		aclChanges = append(aclChanges, change)
	}

	return a.AddChanges(ctx, aclChanges...)
}

func (a *aclTree) AddChanges(ctx context.Context, changes ...*Change) (AddResult, error) {
	a.Lock()
	// TODO: make proper error handling, because there are a lot of corner cases where this will break
	var err error
	var mode Mode

	defer func() {
		if err != nil {
			return
		}

		err = a.removeOrphans()
		if err != nil {
			return
		}

		err = a.treeStorage.SetHeads(a.fullTree.Heads())
		if err != nil {
			return
		}

		a.Unlock()
		switch mode {
		case Append:
			a.updateListener.Update(a)
		case Rebuild:
			a.updateListener.Rebuild(a)
		default:
			break
		}
	}()

	for _, ch := range changes {
		err = a.treeStorage.AddChange(ch)
		if err != nil {
			return AddResult{}, err
		}
		err = a.treeStorage.AddOrphans(ch.Id)
		if err != nil {
			return AddResult{}, err
		}
	}

	prevHeads := a.fullTree.Heads()
	mode = a.fullTree.Add(changes...)
	switch mode {
	case Nothing:
		return AddResult{
			OldHeads: prevHeads,
			Heads:    prevHeads,
			Summary:  AddResultSummaryNothing,
		}, nil

	case Rebuild:
		err = a.rebuildFromStorage(false)
		if err != nil {
			return AddResult{}, err
		}

		return AddResult{
			OldHeads: prevHeads,
			Heads:    a.fullTree.Heads(),
			Summary:  AddResultSummaryRebuild,
		}, nil
	default:
		// just rebuilding the state from start without reloading everything from tree storage
		// as an optimization we could've started from current heads, but I didn't implement that
		a.aclState, err = a.aclStateBuilder.Build()
		if err != nil {
			return AddResult{}, err
		}

		return AddResult{
			OldHeads: prevHeads,
			Heads:    a.fullTree.Heads(),
			Summary:  AddResultSummaryAppend,
		}, nil
	}
}

func (a *aclTree) Iterate(f func(change *Change) bool) {
	a.RLock()
	defer a.RUnlock()
	a.fullTree.Iterate(a.fullTree.RootId(), f)
}

func (a *aclTree) IterateFrom(s string, f func(change *Change) bool) {
	a.RLock()
	defer a.RUnlock()
	a.fullTree.Iterate(s, f)
}

func (a *aclTree) HasChange(s string) bool {
	a.RLock()
	defer a.RUnlock()
	_, attachedExists := a.fullTree.attached[s]
	_, unattachedExists := a.fullTree.unAttached[s]
	_, invalidExists := a.fullTree.invalidChanges[s]
	return attachedExists || unattachedExists || invalidExists
}

func (a *aclTree) Heads() []string {
	a.RLock()
	defer a.RUnlock()
	return a.fullTree.Heads()
}

func (a *aclTree) Root() *Change {
	a.RLock()
	defer a.RUnlock()
	return a.fullTree.Root()
}

func (a *aclTree) Close() error {
	return nil
}
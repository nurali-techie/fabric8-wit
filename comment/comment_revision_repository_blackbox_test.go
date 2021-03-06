package comment_test

import (
	"context"
	"testing"

	"github.com/fabric8-services/fabric8-common/id"

	"github.com/fabric8-services/fabric8-wit/comment"
	"github.com/fabric8-services/fabric8-wit/gormtestsupport"
	"github.com/fabric8-services/fabric8-wit/rendering"
	"github.com/fabric8-services/fabric8-wit/resource"
	tf "github.com/fabric8-services/fabric8-wit/test/testfixture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestRunCommentRevisionRepositoryBlackBoxTest(t *testing.T) {
	resource.Require(t, resource.Database)
	suite.Run(t, &revisionRepositoryBlackBoxTest{DBTestSuite: gormtestsupport.NewDBTestSuite()})
}

type revisionRepositoryBlackBoxTest struct {
	gormtestsupport.DBTestSuite
}

func (s *revisionRepositoryBlackBoxTest) TestStoreCommentRevisions() {
	// given a comment
	fxt := tf.NewTestFixture(s.T(), s.DB, tf.Identities(3), tf.Comments(1))
	c := *fxt.Comments[0]
	// modify the comment
	repository := comment.NewRepository(s.DB)
	revisionRepository := comment.NewRevisionRepository(s.DB)
	c.Body = "Updated body"
	c.Markup = rendering.SystemMarkupPlainText
	err := repository.Save(context.Background(), &c, fxt.Identities[1].ID)
	require.NoError(s.T(), err)
	// modify again the comment
	c.Body = "Updated body2"
	c.Markup = rendering.SystemMarkupMarkdown
	err = repository.Save(context.Background(), &c, fxt.Identities[1].ID)
	require.NoError(s.T(), err)
	// delete the comment
	err = repository.Delete(context.Background(), c.ID, fxt.Identities[2].ID)
	require.NoError(s.T(), err)
	// when
	commentRevisions, err := revisionRepository.List(context.Background(), c.ID)
	// then
	require.NoError(s.T(), err)
	require.Len(s.T(), commentRevisions, 4)
	// revision 1
	revision1 := commentRevisions[0]
	assert.Equal(s.T(), c.ID, revision1.CommentID)
	assert.Equal(s.T(), c.ParentID, revision1.CommentParentID)
	assert.Equal(s.T(), comment.RevisionTypeCreate, revision1.Type)
	assert.Equal(s.T(), fxt.Comments[0].Body, *revision1.CommentBody)
	assert.Equal(s.T(), rendering.SystemMarkupMarkdown, *revision1.CommentMarkup)
	assert.Equal(s.T(), fxt.Identities[0].ID, revision1.ModifierIdentity)
	// revision 2
	revision2 := commentRevisions[1]
	assert.Equal(s.T(), c.ID, revision2.CommentID)
	assert.Equal(s.T(), c.ParentID, revision2.CommentParentID)
	assert.Equal(s.T(), comment.RevisionTypeUpdate, revision2.Type)
	assert.Equal(s.T(), "Updated body", *revision2.CommentBody)
	assert.Equal(s.T(), rendering.SystemMarkupPlainText, *revision2.CommentMarkup)
	assert.Equal(s.T(), fxt.Identities[1].ID, revision2.ModifierIdentity)
	// revision 3
	revision3 := commentRevisions[2]
	assert.Equal(s.T(), c.ID, revision3.CommentID)
	assert.Equal(s.T(), c.ParentID, revision3.CommentParentID)
	assert.Equal(s.T(), comment.RevisionTypeUpdate, revision3.Type)
	assert.Equal(s.T(), "Updated body2", *revision3.CommentBody)
	assert.Equal(s.T(), rendering.SystemMarkupMarkdown, *revision3.CommentMarkup)
	assert.Equal(s.T(), fxt.Identities[1].ID, revision3.ModifierIdentity)
	// revision 4
	revision4 := commentRevisions[3]
	assert.Equal(s.T(), c.ID, revision4.CommentID)
	assert.Equal(s.T(), c.ParentID, revision4.CommentParentID)
	assert.Equal(s.T(), comment.RevisionTypeDelete, revision4.Type)
	assert.Nil(s.T(), revision4.CommentBody)
	assert.Nil(s.T(), revision4.CommentMarkup)
	assert.Equal(s.T(), fxt.Identities[2].ID, revision4.ModifierIdentity)
}

func (s *revisionRepositoryBlackBoxTest) TestStoreChildCommentRevisions() {
	// given a comment
	fxt := tf.NewTestFixture(s.T(), s.DB, tf.Identities(3), tf.Comments(2))
	p := *fxt.Comments[0]
	// make sure the parent comment ID in the parent is nil
	p.ParentCommentID.Valid = false
	// store the parent comment
	repository := comment.NewRepository(s.DB)
	revisionRepository := comment.NewRevisionRepository(s.DB)
	err := repository.Save(context.Background(), &p, fxt.Identities[1].ID)
	require.NoError(s.T(), err)
	parentCommentID := p.ID
	// now create a child comment
	c := *fxt.Comments[1]
	c.Body = "Updated body"
	c.Markup = rendering.SystemMarkupPlainText
	c.ParentCommentID = id.NullUUID{
		UUID:  parentCommentID,
		Valid: true,
	}
	err = repository.Save(context.Background(), &c, fxt.Identities[1].ID)
	require.NoError(s.T(), err)
	// modify again the comment
	c.Body = "Updated body2"
	c.Markup = rendering.SystemMarkupMarkdown
	err = repository.Save(context.Background(), &c, fxt.Identities[1].ID)
	require.NoError(s.T(), err)
	// delete the comment
	err = repository.Delete(context.Background(), c.ID, fxt.Identities[2].ID)
	require.NoError(s.T(), err)
	// when
	commentRevisions, err := revisionRepository.List(context.Background(), c.ID)
	// then
	require.NoError(s.T(), err)
	require.Len(s.T(), commentRevisions, 4)
	// revision 1 - initial create by the fixture
	revision1 := commentRevisions[0]
	assert.Equal(s.T(), c.ID, revision1.CommentID)
	assert.Equal(s.T(), c.ParentID, revision1.CommentParentID)
	assert.Equal(s.T(), comment.RevisionTypeCreate, revision1.Type)
	assert.Equal(s.T(), fxt.Comments[1].Body, *revision1.CommentBody)
	assert.Equal(s.T(), rendering.SystemMarkupMarkdown, *revision1.CommentMarkup)
	assert.Equal(s.T(), fxt.Identities[0].ID, revision1.ModifierIdentity)
	// this is the initial revision created by the fixture, the comment parent id should be nil
	assert.False(s.T(), revision1.CommentParentCommentID.Valid)
	// revision 2 - first create from this test
	revision2 := commentRevisions[1]
	assert.Equal(s.T(), c.ID, revision2.CommentID)
	assert.Equal(s.T(), c.ParentID, revision2.CommentParentID)
	assert.Equal(s.T(), comment.RevisionTypeUpdate, revision2.Type)
	assert.Equal(s.T(), "Updated body", *revision2.CommentBody)
	assert.Equal(s.T(), rendering.SystemMarkupPlainText, *revision2.CommentMarkup)
	assert.Equal(s.T(), fxt.Identities[1].ID, revision2.ModifierIdentity)
	// check if the parent comment id is in the revision
	assert.Equal(s.T(), parentCommentID.String(), revision2.CommentParentCommentID.UUID.String())
	// revision 3 - update from this test
	revision3 := commentRevisions[2]
	assert.Equal(s.T(), c.ID, revision3.CommentID)
	assert.Equal(s.T(), c.ParentID, revision3.CommentParentID)
	assert.Equal(s.T(), comment.RevisionTypeUpdate, revision3.Type)
	assert.Equal(s.T(), "Updated body2", *revision3.CommentBody)
	assert.Equal(s.T(), rendering.SystemMarkupMarkdown, *revision3.CommentMarkup)
	assert.Equal(s.T(), fxt.Identities[1].ID, revision3.ModifierIdentity)
	// check if the parent comment id is in the revision
	assert.Equal(s.T(), parentCommentID.String(), revision3.CommentParentCommentID.UUID.String())
	// revision 4 - delete from this test
	revision4 := commentRevisions[3]
	assert.Equal(s.T(), c.ID, revision4.CommentID)
	assert.Equal(s.T(), c.ParentID, revision4.CommentParentID)
	assert.Equal(s.T(), comment.RevisionTypeDelete, revision4.Type)
	assert.Nil(s.T(), revision4.CommentBody)
	assert.Nil(s.T(), revision4.CommentMarkup)
	assert.Equal(s.T(), fxt.Identities[2].ID, revision4.ModifierIdentity)
	// check if the parent comment id is in the revision
	assert.Equal(s.T(), parentCommentID.String(), revision4.CommentParentCommentID.UUID.String())
}

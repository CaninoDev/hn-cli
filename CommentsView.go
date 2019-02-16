package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/caninodev/hackernewsterm/hnapi"
	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var (
	treeNextSlide func()
)

// Adapted from github.com/johnshiver/plankton/terminal/treeview.go
func createChildrenCommentNodes(rootCommentItem *hnapi.Item) *tview.TreeNode {
	var addChildNode func(commentItem *hnapi.Item) *tview.TreeNode
	addChildNode = func(commentItem *hnapi.Item) *tview.TreeNode {
		tm := time.Unix(commentItem.Time, 0)
		commentText := fmt.Sprintf("[-:-:-]%s[::d] (%s) %d replies", commentItem.By, humanize.Time(tm), len(commentItem.Kids))
		commentNode := tview.NewTreeNode(commentText)
		commentNode.SetReference(commentItem)

		for _, kidID := range commentItem.Kids {
			kid, err := app.api.GetItem(kidID)
			if err != nil {
				log.Print(err)
			}
			cNode := addChildNode(kid)
			if len(kid.Kids) > 0 {
				cNode.SetColor(tcell.ColorMediumSeaGreen)
				cNode.SetSelectable(true)
				cNode.SetSelectedFunc(func() {
					cNode.SetExpanded(!cNode.IsExpanded())
				})
			}
			commentNode.AddChild(cNode)
		}
		return commentNode
	}

	return addChildNode(rootCommentItem)
}

// Comments is a page that will render the comments tree as well as the selected comments
func Comments(nextSlide func()) (title string, content tview.Primitive) {
	treeNextSlide = nextSlide

	app.gui.commentTitle = tview.NewTextView()
	app.gui.commentTitle.SetBorderPadding(1,0,0,0)

	app.gui.commentsContent = tview.NewTextView()
	app.gui.commentsContent.SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true).
		SetWordWrap(true).
		SetBorderPadding(2, 0, 5, 5)

	placeholderNode := tview.NewTreeNode("")
	app.gui.comments = tview.NewTreeView().
		SetGraphics(true).
		// SetChangedFunc(func(n *tview.TreeNode) {
		// 	app.gui.commentsContent.Clear()
		// 	item := n.GetReference().(*hnapi.Item)
		// 	str := html.UnescapeString(item.Text)
		// 	if _, err := fmt.Fprint(app.gui.commentsContent, str); err != nil {
		// 		log.Print(err)
		// 	}
		// 	for idx, child := range n.GetChildren() {
		// 		item := child.GetReference().(*hnapi.Item)
		// 		str := strip.StripTags(html.UnescapeString(item.Text))
		// 		for i := 0; i < idx; i++ {
		// 			if _, err := fmt.Fprint(app.gui.commentsContent, " "); err != nil {
		// 				log.Print(err)
		// 			}
		// 		}
		// 		if _, err := fmt.Fprintln(app.gui.commentsContent, str); err != nil {
		// 			log.Print(err)
		// 		}
		// 	}
		// }).
		SetChangedFunc(func(n *tview.TreeNode) {
			item := n.GetReference().(*hnapi.Item)
			var sb strings.Builder
			if _, err := fmt.Fprintf(&sb, "[-:orange:]%s [-:-:d]wrote:[-:-:-]\n%s", item.By, item.Text); err != nil {
				log.Print(err)
			}
			if _, err := app.gui.commentsContent.Write([]byte(sb.String())); err != nil {
				log.Print(err)
			}
		}).
		SetRoot(placeholderNode)

	return "Comments", tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(app.gui.commentTitle, 1, 1, false).
		AddItem(app.gui.comments, 0, 2, true).
		AddItem(app.gui.commentsContent, 0, 5, true)
}

func (gui *GUI) germinate(storyItem hnapi.Item) {
	gui.console.SetText("Loading comments...")
	// var add func(targets *tview.TreeNode) *tview.TreeNode
	add := func(target *tview.TreeNode) *tview.TreeNode {
		for _, rootCommentID := range storyItem.Kids {
			rootComment, err := app.api.GetItem(rootCommentID)
			if err != nil {
				log.Print(err)
			}
			rootCommentNode := createChildrenCommentNodes(rootComment)
			if len(rootComment.Kids) > 0 {
				rootCommentNode.SetSelectable(true).
					SetSelectedFunc(func() {
						rootCommentNode.SetExpanded(!rootCommentNode.IsExpanded())
						rootCommentNode.SetColor(tcell.ColorGreen)
					})
			}
			target.AddChild(rootCommentNode)
		}
		return target
	}

	storyNode := *tview.NewTreeNode(storyItem.Text)
	root := add(&storyNode)
	gui.commentTitle.SetText(storyItem.Title)
	gui.comments.SetRoot(root).
		SetCurrentNode(root).
		SetTopLevel(1)

	gui.console.Clear()
	}

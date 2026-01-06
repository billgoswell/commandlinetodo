package main

import (
	"sort"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type todoList struct {
	id           int
	name         string
	displayOrder int
	archived     bool
	createdAt    int64
	updatedAt    int64
}

type TaskCreationFlowStep int

const (
	TaskFlowInputText TaskCreationFlowStep = iota
	TaskFlowSelectPriority
	TaskFlowSetDueDate
)

// FlowStepToState maps each flow step to its corresponding app state
var FlowStepToState map[TaskCreationFlowStep]AppState

func init() {
	FlowStepToState = map[TaskCreationFlowStep]AppState{
		TaskFlowInputText:      StateTaskInput,
		TaskFlowSelectPriority: StatePrioritySelection,
		TaskFlowSetDueDate:     StateDueDateInput,
	}
}

// TaskCreationFlow manages the multi-step task creation process
type TaskCreationFlow struct {
	step     TaskCreationFlowStep
	text     string
	priority int
	dueDate  int64
}

func newTaskCreationFlow() TaskCreationFlow {
	return TaskCreationFlow{
		step:     TaskFlowInputText,
		text:     "",
		priority: DefaultPriority,
		dueDate:  0,
	}
}

func (f *TaskCreationFlow) nextStep() {
	if f.step < TaskFlowSetDueDate {
		f.step++
	}
}

func (f *TaskCreationFlow) previousStep() {
	if f.step > TaskFlowInputText {
		f.step--
	}
}

func (f *TaskCreationFlow) reset() {
	f.step = TaskFlowInputText
	f.text = ""
	f.priority = DefaultPriority
	f.dueDate = 0
}

// InputContext holds all temporary input/editing state
type InputContext struct {
	itemIndex   int // Index of item being edited (-1 = none)
	deleteIndex int // Index of item pending deletion (-1 = none)
	listIndex   int // Cursor position in list selector
}

func newInputContext() InputContext {
	return InputContext{
		itemIndex:   -1,
		deleteIndex: -1,
		listIndex:   0,
	}
}

type model struct {
	items               []todoItem
	cursor              int
	width               int
	height              int
	textInput           textinput.Model
	viewport            viewport.Model
	scrollOffset        int
	todoLists           []todoList
	currentListID       int
	currentListIndex    int
	errorMsg            string
	filteredItems       []todoItem
	filteredListID      int
	filteredItemIndices []int
	cacheValid          bool
	input               InputContext
	taskFlow            TaskCreationFlow
	currentState        AppState
	currentSubState     SubState
}

func initialModel(todoItems []todoItem, todoLists []todoList) model {
	ti := textinput.New()
	ti.Placeholder = TextInputPlaceholder
	ti.Focus()
	ti.CharLimit = TextInputCharLimit
	ti.Width = TextInputWidth

	vp := viewport.New(ViewportWidth, ViewportHeight)

	currentListID := 0
	currentListIndex := 0
	if len(todoLists) > 0 {
		currentListID = todoLists[0].id
		currentListIndex = 0
	}

	return model{
		items:            todoItems,
		textInput:        ti,
		viewport:         vp,
		scrollOffset:     0,
		todoLists:        todoLists,
		currentListID:    currentListID,
		currentListIndex: currentListIndex,
		input:            newInputContext(),
		taskFlow:         newTaskCreationFlow(),
		currentState:     StateMainBrowse,
		currentSubState:  SubStateNone,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) sortItems() {
	sort.Slice(m.items, func(i, j int) bool {
		if m.items[i].done != m.items[j].done {
			return !m.items[i].done
		}
		return m.items[i].priority < m.items[j].priority
	})
	m.invalidateCache()
}

func (m *model) setState(state AppState, subState SubState) {
	m.currentState = state
	m.currentSubState = subState
	m.viewport.Height = m.getViewportHeight(m.height)
}

func (m *model) returnToMain() {
	m.currentState = StateMainBrowse
	m.currentSubState = SubStateNone
	m.textInput.Reset()
	m.viewport.Height = m.getViewportHeight(m.height)
}

func (m *model) getListAtIndex(index int) *todoList {
	if index < 0 || index >= len(m.todoLists) {
		return nil
	}
	return &m.todoLists[index]
}

func (m *model) getStateForFlowStep(step TaskCreationFlowStep) AppState {
	if state, ok := FlowStepToState[step]; ok {
		return state
	}
	return StateMainBrowse
}

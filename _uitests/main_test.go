package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/mxschmitt/playwright-go"
)

const (
	url              = "http://localhost:8089/chat"
	nameInput        = `//input[@id="name"]`
	joinButton       = `//input[@id="join"]`
	msgOuputDiv      = `//div[@id="messages"]`
	msgInputTextarea = `//textarea[@id="msg"]`
	sendInput        = `//input[@id="send"]`
	nameSpan         = `//span[contains(@class, "name") and text()="%s"]`
	nameColorSpan    = `//span[contains(@class, "name") and text()="%s" and @style="color:%s"]`
	msgSpan          = `//span[contains(@class, "msg") and text()="%s"]`
	meCmdSpan        = `//span[contains(@class, "cmdme") and text()="%s"]`
	commandSpan      = `//span[contains(@class, "command")]`
)

var pw *playwright.Playwright

func openChat(t *testing.T, browser playwright.Browser, username string) (playwright.Page, error) {
	page, err := browser.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not open new browser page: %w", err)
	}

	_, err = page.Goto(url)
	if err != nil {
		return nil, fmt.Errorf("could not open chat page: %w", err)
	}

	_, err = page.WaitForSelector(nameInput)
	if err != nil {
		return nil, fmt.Errorf("could not find name input: %w", err)
	}

	err = page.Type(nameInput, username)
	if err != nil {
		return nil, fmt.Errorf("an error occured when typing the username: %w", err)
	}

	err = page.Click(joinButton)
	if err != nil {
		return nil, fmt.Errorf("could not click join button: %w", err)
	}

	_, err = page.WaitForSelector(msgOuputDiv)
	if err != nil {
		return nil, fmt.Errorf("chat window did not show: %w", err)
	}

	_, err = page.WaitForSelector(fmt.Sprintf(nameSpan, username))
	if err != nil {
		return nil, fmt.Errorf("could not get join message span: %w", err)
	}

	return page, nil
}

func sendMessage(page playwright.Page, msg string) error {
	err := page.Type(msgInputTextarea, msg)
	if err != nil {
		return fmt.Errorf("could not find msg box: %w", err)
	}

	err = page.Click(sendInput)
	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	return nil
}

func TestMain(m *testing.M) {
	var err error
	pw, err = playwright.Run()
	if err != nil {
		fmt.Printf("could not run playwright: %v", err)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

func TestAccess(t *testing.T) {
	browser, err := pw.Firefox.Launch()
	if err != nil {
		t.Errorf("could not launch browser: %v", err)
	}
	defer browser.Close()

	_, err = openChat(t, browser, "testUser")
	if err != nil {
		t.Error(err)
	}
}

func TestChatMessage(t *testing.T) {
	const msg = "testing 1 2 3"

	browser, err := pw.Firefox.Launch()
	if err != nil {
		t.Errorf("could not launch browser: %v", err)
	}
	defer browser.Close()

	page1, err := openChat(t, browser, "testUser1")
	if err != nil {
		t.Errorf("could not open chat for user 1: %v", err)
	}

	page2, err := openChat(t, browser, "testUser2")
	if err != nil {
		t.Errorf("could not open chat for user 2: %v", err)
	}

	err = sendMessage(page1, msg)
	if err != nil {
		t.Error(err)
	}

	_, err = page2.WaitForSelector(fmt.Sprintf(msgSpan, msg))
	if err != nil {
		t.Errorf("did not find testing msg: %v", err)
	}
}

func TestCommandNick(t *testing.T) {
	const nick = "newNick"

	browser, err := pw.Firefox.Launch()
	if err != nil {
		t.Errorf("could not launch browser: %v", err)
	}
	defer browser.Close()

	page, err := openChat(t, browser, "testUser")
	if err != nil {
		t.Error(err)
	}

	err = sendMessage(page, fmt.Sprintf("/nick %s", nick))
	if err != nil {
		t.Error(err)
	}

	_, err = page.WaitForSelector(fmt.Sprintf(nameSpan, nick))
	if err != nil {
		t.Errorf("could not find nick change message: %v", err)
	}
}

func TestCommandMe(t *testing.T) {
	const msg = "this is a me message"

	browser, err := pw.Firefox.Launch()
	if err != nil {
		t.Errorf("could not launch browser: %v", err)
	}
	defer browser.Close()

	page, err := openChat(t, browser, "testUser")
	if err != nil {
		t.Error(err)
	}

	err = sendMessage(page, fmt.Sprintf("/me %s", msg))
	if err != nil {
		t.Error(err)
	}

	_, err = page.WaitForSelector(fmt.Sprintf(meCmdSpan, msg))
	if err != nil {
		t.Errorf("could not find user me message: %v", err)
	}
}

func TestCommandColor(t *testing.T) {
	const (
		name  = "testUser"
		color = "red"
	)

	browser, err := pw.Firefox.Launch()
	if err != nil {
		t.Errorf("could not launch browser: %v", err)
	}
	defer browser.Close()

	page, err := openChat(t, browser, name)
	if err != nil {
		t.Error(err)
	}

	err = sendMessage(page, fmt.Sprintf("/color %s", color))
	if err != nil {
		t.Errorf("failed to send /color command: %v", err)
	}

	_, err = page.WaitForSelector(`//span[contains(@class, "command") and contains(text(),"Color changed successfully")]`)
	if err != nil {
		t.Errorf("could not find color change notification: %v", err)
	}

	err = sendMessage(page, "test")
	if err != nil {
		t.Error(err)
	}

	_, err = page.WaitForSelector(fmt.Sprintf(nameColorSpan, name, color))
	if err != nil {
		t.Errorf("could not find user message with new color: %v", err)
	}
}

func TestGenericCommands(t *testing.T) {
	wrapFunc := func(command string) func(*testing.T) {
		return func(t *testing.T) {
			browser, err := pw.Firefox.Launch()
			if err != nil {
				t.Errorf("could not launch browser: %v", err)
			}
			defer browser.Close()

			page, err := openChat(t, browser, "testUser")
			if err != nil {
				t.Error(err)
			}

			err = sendMessage(page, command)
			if err != nil {
				t.Errorf("failed to send %s command: %v", command, err)
			}

			_, err = page.WaitForSelector(commandSpan)
			if err != nil {
				t.Errorf("could not find command message: %v", err)
			}
		}
	}

	t.Run("pin", wrapFunc("/pin"))
	t.Run("stats", wrapFunc("/stats"))
	t.Run("users", wrapFunc("/users"))
	t.Run("whoami", wrapFunc("/whoami"))
}

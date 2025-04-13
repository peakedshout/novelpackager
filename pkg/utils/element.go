package utils

import (
	"github.com/go-rod/rod"
	"github.com/peakedshout/novelpackager/pkg/model"
	"strings"
)

func ElementType(element *rod.Element) string {
	split := strings.Split(element.MustGetXPath(true), "/")
	cut, _, b := strings.Cut(split[len(split)-1], "[")
	if !b {
		return ""
	}
	return cut
}

type Elementor interface {
	Element(selector string) (*rod.Element, error)
	Elements(selector string) (rod.Elements, error)
	ElementX(selector string) (*rod.Element, error)
	ElementsX(selector string) (rod.Elements, error)
}

type Element string

func (e Element) Text(er Elementor) (string, error) {
	el, err := e.Element(er)
	if err != nil {
		return "", err
	}
	text, err := el.Text()
	if err != nil {
		return "", err
	}
	return text, nil
}

func (e Element) Resource(er Elementor) ([]byte, error) {
	el, err := e.Element(er)
	if err != nil {
		return nil, err
	}
	bs, err := el.Resource()
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (e Element) Attribute(er Elementor, name string) (*string, error) {
	el, err := e.Element(er)
	if err != nil {
		return nil, err
	}
	attr, err := el.Attribute(name)
	if err != nil {
		return nil, err
	}
	return attr, nil
}

func (e Element) Element(er Elementor) (*rod.Element, error) {
	el, err := er.Element(string(e))
	if err != nil {
		return nil, err
	}
	return el, nil
}

func (e Element) Elements(er Elementor) (rod.Elements, error) {
	el, err := er.Elements(string(e))
	if err != nil {
		return nil, model.ErrElement.Errorf(e, err)
	}
	return el, nil
}

type ElementX string

func (e ElementX) Text(er Elementor) (string, error) {
	el, err := e.ElementX(er)
	if err != nil {
		return "", err
	}
	text, err := el.Text()
	if err != nil {
		return "", err
	}
	return text, nil
}

func (e ElementX) Resource(er Elementor) ([]byte, error) {
	el, err := e.ElementX(er)
	if err != nil {
		return nil, err
	}
	bs, err := el.Resource()
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (e ElementX) ElementX(er Elementor) (*rod.Element, error) {
	el, err := er.ElementX(string(e))
	if err != nil {
		return nil, err
	}
	return el, nil
}

func (e ElementX) ElementsX(er Elementor) (rod.Elements, error) {
	el, err := er.ElementsX(string(e))
	if err != nil {
		return nil, err
	}
	return el, nil
}

func (e ElementX) Attribute(er Elementor, name string) (*string, error) {
	el, err := e.ElementX(er)
	if err != nil {
		return nil, err
	}
	attr, err := el.Attribute(name)
	if err != nil {
		return nil, err
	}
	return attr, nil
}

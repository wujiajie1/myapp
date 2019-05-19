package retrier

import (
	"errors"
	"testing"
	"vendor"
)

var (
	errFoo = errors.New("FOO")
	errBar = errors.New("BAR")
	errBaz = errors.New("BAZ")
)

func TestDefaultClassifier(t *testing.T) {
	c := vendor.DefaultClassifier{}

	if c.Classify(nil) != vendor.Succeed {
		t.Error("default misclassified nil")
	}

	if c.Classify(errFoo) != vendor.Retry {
		t.Error("default misclassified foo")
	}
	if c.Classify(errBar) != vendor.Retry {
		t.Error("default misclassified bar")
	}
	if c.Classify(errBaz) != vendor.Retry {
		t.Error("default misclassified baz")
	}
}

func TestWhitelistClassifier(t *testing.T) {
	c := vendor.WhitelistClassifier{errFoo, errBar}

	if c.Classify(nil) != vendor.Succeed {
		t.Error("whitelist misclassified nil")
	}

	if c.Classify(errFoo) != vendor.Retry {
		t.Error("whitelist misclassified foo")
	}
	if c.Classify(errBar) != vendor.Retry {
		t.Error("whitelist misclassified bar")
	}
	if c.Classify(errBaz) != vendor.Fail {
		t.Error("whitelist misclassified baz")
	}
}

func TestBlacklistClassifier(t *testing.T) {
	c := vendor.BlacklistClassifier{errBar}

	if c.Classify(nil) != vendor.Succeed {
		t.Error("blacklist misclassified nil")
	}

	if c.Classify(errFoo) != vendor.Retry {
		t.Error("blacklist misclassified foo")
	}
	if c.Classify(errBar) != vendor.Fail {
		t.Error("blacklist misclassified bar")
	}
	if c.Classify(errBaz) != vendor.Retry {
		t.Error("blacklist misclassified baz")
	}
}

package main

/* stdlib includes */
import (
	"flag"
	"fmt"
	"os"
	"strings"
)

/* external includes */
import "github.com/op/go-logging"
import "github.com/mitchellh/go-libucl"

/* definitions */

/* ConfigLevelType is how far we are nested into the config */
type ConfigLevelType int

const (
	ConfigLevelRoot ConfigLevelType = iota
	ConfigLevelGroup
	ConfigLevelRule
)

type RuleStateType int

const (
	RuleStateUnknown RuleStateType = iota
	RuleStateSuccess
	RuleStateFail
)

type RuleStatusType int

const (
	RuleStatusEnabled RuleStatusType = iota
	/* a disabled service leaves the run-time configuration */
	RuleStatusDisabled

	/* run the rule once at startup, then disable the rule
	 * helpful for failing over hosts, or services
	 */
	RuleStatusRunOnce
	RuleStatusRunOnceFail
	RuleStatusRunOnceSuccess

	/* are these helpful, at all? */
	RuleStatusAlwaysFail
	RuleStatusAlwaysSuccess
)

type Rule struct {
	/* name of the grouping for the rule */
	groupName string

	/* name of the rule in the grouping */
	name string

	/* what is the status of this rule */
	status RuleStatusType

	/* what is the period between scheduled runs */
	interval float64

	/* what is the period between scheduled runs on previously failed rules */
	failInterval float64

	/* shell command to run to initiate test */
	/*  hoping to extend to support go-native tests */
	test string

	/* shell command to run when the state changes to failed */
	changeFail string

	/* shell command to run when the state changes to success */
	changeSuccess string

	/* the result of the last rule check */
	lastState RuleStateType
}

/* meat */

/* dependancy injection is for another day */
var log = logging.MustGetLogger(os.Args[0])

var rulesDefaults = make(map[string]Rule)
var rules = make(map[string]Rule)

func loadConfiguration(configPath string) (*libucl.Object, error) {
	p := libucl.NewParser(0)
	defer p.Close()

	e := p.AddFile(configPath)
	if e != nil {
		log.Error(fmt.Sprintf("Could not load configuration file %v: %+v", configPath, e))
		return nil, e
	}

	config := p.Object()
	return config, nil
}

/* config format:
 *  default
 *  group
 *    rule
 *  group
 *    rule
 *    rule
 *  rule
 *  default
 */
func walkConfiguration(config *libucl.Object, parentRule string, depth ConfigLevelType) {
	var name string
	if depth == ConfigLevelRoot {
		name = "default"
	} else {
		if parentRule == "default" {
			name = config.Key()
		} else {
			name = parentRule + "/" + config.Key()
		}
	}

	if name == "" {
		/* XXX: if name has already been assigned push warning up the stack */
		return
	} else if _, ok := rules[name]; ok {
		/* XXX: if name has already been assigned push warning up the stack */
		return
	} else if _, ok := rulesDefaults[name]; ok {
		/* XXX: if name has already been assigned push warning up the stack */
		return
	}

	var nextDepth ConfigLevelType
	//	tabs := 0
	isRule := (config.Get("test") != nil)

	switch depth {
	case ConfigLevelRoot:
		nextDepth = ConfigLevelGroup
	case ConfigLevelGroup:
		//		tabs = 1
		nextDepth = ConfigLevelRule
	case ConfigLevelRule:
		if !isRule {
			/* XXX: can only define rules at this level */
			return
		}
		//		tabs = 2
	}

	rule := Rule{name: name, groupName: parentRule}

	i := config.Iterate(true)
	defer i.Close()

	for c := i.Next(); c != nil; c = i.Next() {
		defer c.Close()
		//fmt.Printf("%v %v %v %v\n", c.Key(), c.Type(), depth, isRule)

		if c.Type() == libucl.ObjectTypeObject {
			/* if we are a rule, we stop parsing children */
			if depth != ConfigLevelRule || !isRule {
				//fmt.Printf("%s%v: \n", strings.Repeat("\t", tabs), c.Key())
				walkConfiguration(c, name, nextDepth)
			} else {
				/* XXX: push warning up the stack */
			}

			continue
		}

		switch strings.ToLower(c.Key()) {
		case "status":
			if c.Type() != libucl.ObjectTypeString {
				/* XXX: push warning up the stack */
			}

			switch strings.ToLower(c.ToString()) {
			case "enabled":
				rule.status = RuleStatusEnabled
			case "disabled":
				rule.status = RuleStatusDisabled
			case "run-once":
				rule.status = RuleStatusRunOnce
			case "run-once-fail":
				rule.status = RuleStatusRunOnceFail
			case "run-once-sucess":
				rule.status = RuleStatusRunOnceSuccess
			case "always-fail":
				rule.status = RuleStatusAlwaysFail
			case "always-success":
				rule.status = RuleStatusAlwaysSuccess
			default:
				/* XXX: push warning up the stack */
			}
		case "interval":
			switch c.Type() {
			case libucl.ObjectTypeInt, libucl.ObjectTypeFloat, libucl.ObjectTypeTime:
				rule.interval = c.ToFloat()
			default:
				/* XXX: push warning up the stack */
			}
		case "fail_interval":
			switch c.Type() {
			case libucl.ObjectTypeInt, libucl.ObjectTypeFloat, libucl.ObjectTypeTime:
				rule.interval = c.ToFloat()
			default:
				/* XXX: push warning up the stack */
			}
		case "test":
			if c.Type() != libucl.ObjectTypeString {
				/* XXX: push warning up the stack */
			}
			rule.test = c.ToString()
		case "change_fail":
			if c.Type() != libucl.ObjectTypeString {
				/* XXX: push warning up the stack */
			}
			rule.changeFail = c.ToString()
		case "change_success":
			if c.Type() != libucl.ObjectTypeString {
				/* XXX: push warning up the stack */
			}
			rule.changeSuccess = c.ToString()
		default:
			//fmt.Printf("%s%+v\n", strings.Repeat("\t", tabs), c)
			/* XXX: push warning up the stack */
		}
	}

	//fmt.Printf("%s%+v\n", strings.Repeat("\t", tabs), rule)

	if !isRule {
		rulesDefaults[name] = rule
		return
	}
	rules[name] = rule

	return
}

func main() {
	var configPath string

	flag.StringVar(&configPath, "config", "etc/hfm.conf", "Configuration file path")
	flag.Parse()

	config, e := loadConfiguration(configPath)
	if e != nil {
		log.Error(fmt.Sprintf("Could not load configuration file %v: %+v", configPath, e))
		panic(e)
	}
	//	fmt.Println(config.Emit(libucl.EmitConfig))

	fmt.Println("Building ruleset...")
	walkConfiguration(config, "", ConfigLevelRoot)
	fmt.Println("end...")

	fmt.Println("Rule defaults")
	for _, rule := range rulesDefaults {
		fmt.Printf("%+v\n", rule)
	}

	fmt.Println("")
	fmt.Println("Rules")
	for _, rule := range rules {
		fmt.Printf("%+v\n", rule)
	}

}

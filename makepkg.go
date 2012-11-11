package xsd

import (
	"fmt"
	"strings"

	util "github.com/metaleap/go-util"
	ustr "github.com/metaleap/go-util/str"

	xsdt "github.com/metaleap/go-xsd/types"
)

var (
	PkgGen = &pkgGen {
		BaseCodePath: util.BaseCodePath("metaleap", "go-xsd-pkg"),
		BasePath: "github.com/metaleap/go-xsd-pkg",
		ForceParseForDefaults: false,
	}
	perPkgState struct {
		anonCounts map[string]uint64
		attsCache, elemsCacheOnce, elemsCacheMult map[string]string
		attGroups, attGroupRefImps map[*AttributeGroup]string
		attsKeys, attRefImps map[*Attribute]string
		elemGroups, elemGroupRefImps map[*Group]string
		elemChoices, elemChoiceRefImps map[*Choice]string
		elemSeqs, elemSeqRefImps map[*Sequence]string
		elemKeys, elemRefImps map[*Element]string
		simpleContentValueTypes map[string]string
	}
)

type pkgGen struct {
	BaseCodePath, BasePath string
	ForceParseForDefaults bool
}

type beforeAfterMake interface {
	afterMakePkg (*PkgBag)
	beforeMakePkg (*PkgBag)
}

type pkgStack []interface{}

	func (me *pkgStack) Pop () (el interface{}) { sl := *me; el = sl[0]; *me = sl[1 :]; return }

	func (me *pkgStack) Push (el interface{}) { nu := []interface{} { el }; *me = append(nu, *me...) }

type pkgStacks struct {
	Name, SimpleType pkgStack
}

	func (me *pkgStacks) CurName () (r xsdt.NCName) {
		if len(me.Name) > 0 { r = me.Name[0].(xsdt.NCName) }; return
	}

	func (me *pkgStacks) CurSimpleType () (r *SimpleType) {
		if len(me.SimpleType) > 0 { r = me.SimpleType[0].(*SimpleType) }; return
	}

type PkgBag struct {
	Schema *Schema
	Stacks pkgStacks
	ParseTypes []string

	lines []string
	impName string
	imports map[string]string
	impsUsed map[string]bool
	now int64
	snow string
}

	func (me *PkgBag) AnonName (n string) (an xsdt.NCName) {
		var c uint64
		n = "Txsd" + n
		an = xsdt.NCName(n)
		if c = perPkgState.anonCounts[n]; c > 0 { an += xsdt.NCName(fmt.Sprintf("%v", c)) }
		perPkgState.anonCounts[n] = c + 1
		return
	}

	func (me *PkgBag) append (lines ... string) {
		me.lines = append(me.lines, lines ...)
	}

	func (me *PkgBag) appendFmt (addLineAfter bool, format string, fmtArgs ... interface{}) {
		me.append(fmt.Sprintf(format, fmtArgs ...))
		if addLineAfter { me.append("") }
	}

	func (me *PkgBag) insertFmt (index int, format string, fmtArgs ... interface{}) {
		me.lines = append(me.lines[: index], append([]string { fmt.Sprintf(format, fmtArgs ...) }, me.lines[index : ] ...) ...)
	}

	func (me *PkgBag) isParseType (typeRef string) (bool) {
		for _, pt := range me.ParseTypes { if typeRef == pt { return true } }
		return false
	}

	func (me *PkgBag) reinit () {
		me.impName = "xsdt"
		me.imports, me.impsUsed, me.lines = map[string]string {}, map[string]bool {}, []string { "//\tAuto-generated by the \"go-xsd\" package located at:", "//\t\tgithub.com/metaleap/go-xsd", "//\tComments on types and fields (if any) are from the XSD file located at:", "//\t\t" + me.Schema.loadUri, "package gopkg_" + me.safeName(me.Schema.RootSchema().loadUri), "" }
	}

	func (me *PkgBag) resolveQnameRef (ref, pref string, noUsageRec *string) string {
		var ns = me.Schema.XMLNamespaces[""]
		var impName = ""
		if len(ref) == 0 { return "" }
		if pos := strings.Index(ref, ":"); pos > 0 {
			impName, ns = ref[: pos], me.Schema.XMLNamespaces[ref[: pos]]
			ref = ref[(pos + 1) :]
		}
		if ns == xsdNamespaceUri { impName, pref = me.impName, "" }
		if ns == me.Schema.TargetNamespace.String() { impName = "" }
		if noUsageRec == nil { me.impsUsed[impName] = true } else { *noUsageRec = impName }
		return ustr.PrefixWithSep(impName, ".", me.safeName(ustr.PrependIf(ref, pref)))
	}

	func (me *PkgBag) safeName (name string) string {
		return ustr.SafeIdentifier(name)
	}

	func (me *PkgBag) xsdStringTypeRef () string {
		return ustr.PrefixWithSep(me.Schema.XSDNamespace, ":", "string")
	}

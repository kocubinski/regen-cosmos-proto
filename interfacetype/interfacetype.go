package interfacetype

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/regen-network/cosmos-proto"
	"strings"
)

type interfacetype struct {
	*generator.Generator
	generator.PluginImports
}

func NewInterfaceType() *interfacetype {
	return &interfacetype{}
}

func (p *interfacetype) Name() string {
	return "interfacetype"
}

func (p *interfacetype) Init(g *generator.Generator) {
	p.Generator = g
}

func GetInterfaceType(message *descriptor.DescriptorProto) string {
	if message == nil {
		return ""
	}
	if message.Options != nil {
		v, err := proto.GetExtension(message.Options, cosmos_proto.E_InterfaceType)
		if err == nil && v.(*string) != nil {
			return *(v.(*string))
		}
	}
	return ""
}

func (p *interfacetype) Generate(file *generator.FileDescriptor) {
	p.PluginImports = generator.NewPluginImports(p.Generator)

	for _, message := range file.Messages() {
		iface := GetInterfaceType(message.DescriptorProto)
		if len(iface) == 0 {
			continue
		}
		if len(message.OneofDecl) != 1 {
			panic("interfacetype only supports messages with exactly one oneof declaration")
		}
		for _, field := range message.Field {
			if idx := field.OneofIndex; idx == nil || *idx != 0 {
				panic("all fields in interfacetype message must belong to the oneof")
			}
		}

		ifacePackage, ifaceName := splitCPackageType(iface)
		ifaceRef := ifaceName
		if len(ifacePackage) != 0 {
			imp := p.PluginImports.NewImport(ifacePackage).Use()
			ifaceRef = fmt.Sprintf("%s.%s", imp, ifaceName)
		}

		ccTypeName := generator.CamelCaseSlice(message.TypeName())
		p.P(`func (this *`, ccTypeName, `) ToInterface() `, ifaceRef, ` {`)
		p.In()
		for _, field := range message.Field {
			fieldname := p.GetOneOfFieldName(message, field)
			if fieldname == "Value" {
				panic("cannot have a onlyone message " + ccTypeName + " with a field named Value")
			}
			p.P(`if x := this.Get`, fieldname, `(); x != nil {`)
			p.In()
			p.P(`return x`)
			p.Out()
			p.P(`}`)
		}
		p.P(`return nil`)
		p.Out()
		p.P(`}`)
		//p.P(``)
		//p.P(`func (this *`, ccTypeName, `) FromInterface(value `, ifaceName, `) error {`)
		//p.In()
		//p.P(`switch vt := value.(type) {`)
		//p.In()
		//for _, field := range message.Field {
		//	fieldname := p.GetFieldName(message, field)
		//	goTyp, _ := p.GoType(message, field)
		//	p.P(`case `, goTyp, `:`)
		//	p.In()
		//	p.P(`//this.`, fieldname, ` = vt`)
		//	p.Out()
		//}
		//p.P(`default:`)
		//p.In()
		//p.P(`return nil`)
		//p.Out()
		//p.P(`}`)
		//p.Out()
		//p.P(`}`)
	}
}

func splitCPackageType(ctype string) (packageName string, typ string) {
	ss := strings.Split(ctype, ".")
	if len(ss) == 1 {
		return "", ctype
	}
	packageName = strings.Join(ss[0:len(ss)-1], ".")
	typeName := ss[len(ss)-1]
	return packageName, typeName
}

func init() {
	generator.RegisterPlugin(NewInterfaceType())
}

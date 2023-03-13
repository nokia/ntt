package titan

import (
	"encoding/xml"
)

type TITANProjectFileInformation *TopLevelProjectType

type TTCN3preprocessorDefines struct {
	ListItem []string `xml:"listItem"`
}

type TTCN3preprocessorUndefines struct {
	ListItem []string `xml:"listItem"`
}

type PreprocessorDefines struct {
	XMLName  xml.Name `xml:"preprocessorDefines"`
	ListItem []string `xml:"listItem"`
}

type PreprocessorUndefines struct {
	XMLName  xml.Name `xml:"preprocessorUndefines"`
	ListItem []string `xml:"listItem"`
}

type TTCN3preprocessorIncludes struct {
	ListItem []string `xml:"listItem"`
}

type PreprocessorIncludes struct {
	XMLName  xml.Name `xml:"preprocessorIncludes"`
	ListItem []string `xml:"listItem"`
}

type SolarisSpecificLibraries struct {
	ListItem []string `xml:"listItem"`
}

type Solaris8SpecificLibraries struct {
	ListItem []string `xml:"listItem"`
}

type FreeBSDSpecificLibraries struct {
	ListItem []string `xml:"listItem"`
}

type LinuxSpecificLibraries struct {
	ListItem []string `xml:"listItem"`
}

type Win32SpecificLibraries struct {
	ListItem []string `xml:"listItem"`
}

type AdditionalObjects struct {
	XMLName  xml.Name `xml:"additionalObjects"`
	ListItem []string `xml:"listItem"`
}

type LinkerLibraries struct {
	XMLName  xml.Name `xml:"linkerLibraries"`
	ListItem []string `xml:"listItem"`
}

type LinkerLibrarySearchPath struct {
	XMLName  xml.Name `xml:"linkerLibrarySearchPath"`
	ListItem []string `xml:"listItem"`
}

type Target struct {
	NameAttr      string `xml:"name,attr"`
	PlacementAttr string `xml:"placement,attr"`
}

type Targets struct {
	Target []*Target `xml:"Target"`
}

type ProjectSpecificRulesGenerator struct {
	GeneratorCommand string   `xml:"GeneratorCommand"`
	Targets          *Targets `xml:"Targets"`
}

type MakefileSettings struct {
	GenerateMakefile                bool                           `xml:"generateMakefile"`
	GenerateInternalMakefile        bool                           `xml:"generateInternalMakefile"`
	SymboliclinklessBuild           bool                           `xml:"symboliclinklessBuild"`
	UseAbsolutePath                 bool                           `xml:"useAbsolutePath"`
	GNUMake                         bool                           `xml:"GNUMake"`
	IncrementalDependencyRefresh    bool                           `xml:"incrementalDependencyRefresh"`
	DynamicLinking                  bool                           `xml:"dynamicLinking"`
	FunctiontestRuntime             bool                           `xml:"functiontestRuntime"`
	SingleMode                      bool                           `xml:"singleMode"`
	CodeSplitting                   int                            `xml:"codeSplitting"`
	DefaultTarget                   string                         `xml:"defaultTarget"`
	TargetExecutable                string                         `xml:"targetExecutable"`
	TTCN3preprocessor               string                         `xml:"TTCN3preprocessor"`
	TTCN3preprocessorDefines        *TTCN3preprocessorDefines      `xml:"TTCN3preprocessorDefines"`
	TTCN3preprocessorUndefines      *TTCN3preprocessorUndefines    `xml:"TTCN3preprocessorUndefines"`
	PreprocessorDefines             *PreprocessorDefines           `xml:"preprocessorDefines"`
	PreprocessorUndefines           *PreprocessorUndefines         `xml:"preprocessorUndefines"`
	TTCN3preprocessorIncludes       *TTCN3preprocessorIncludes     `xml:"TTCN3preprocessorIncludes"`
	PreprocessorIncludes            *PreprocessorIncludes          `xml:"preprocessorIncludes"`
	SemanticCheckOnly               bool                           `xml:"semanticCheckOnly"`
	DisableAttributeValidation      bool                           `xml:"disableAttributeValidation"`
	DisableBER                      bool                           `xml:"disableBER"`
	DisableRAW                      bool                           `xml:"disableRAW"`
	DisableTEXT                     bool                           `xml:"disableTEXT"`
	DisableXER                      bool                           `xml:"disableXER"`
	DisableJSON                     bool                           `xml:"disableJSON"`
	DisableOER                      bool                           `xml:"disableOER"`
	ForceXERinASN1                  bool                           `xml:"forceXERinASN.1"`
	DefaultasOmit                   bool                           `xml:"defaultasOmit"`
	EnumHackProperty                bool                           `xml:"enumHackProperty"`
	ForceOldFuncOutParHandling      bool                           `xml:"forceOldFuncOutParHandling"`
	GccMessageFormat                bool                           `xml:"gccMessageFormat"`
	LineNumbersOnlyInMessages       bool                           `xml:"lineNumbersOnlyInMessages"`
	IncludeSourceInfo               bool                           `xml:"includeSourceInfo"`
	AddSourceLineInfo               bool                           `xml:"addSourceLineInfo"`
	SuppressWarnings                bool                           `xml:"suppressWarnings"`
	OutParamBoundness               bool                           `xml:"outParamBoundness"`
	OmitInValueList                 bool                           `xml:"omitInValueList"`
	WarningsForBadVariants          bool                           `xml:"warningsForBadVariants"`
	IgnoreUntaggedOnTopLevelUnion   bool                           `xml:"ignoreUntaggedOnTopLevelUnion"`
	ActivateDebugger                bool                           `xml:"activateDebugger"`
	Quietly                         bool                           `xml:"quietly"`
	EnableLegacyEncoding            bool                           `xml:"enableLegacyEncoding"`
	DisableUserInformation          bool                           `xml:"disableUserInformation"`
	EnableRealtimeTesting           bool                           `xml:"enableRealtimeTesting"`
	NamingRules                     string                         `xml:"namingRules"`
	DisableSubtypeChecking          bool                           `xml:"disableSubtypeChecking"`
	ForceGenSeof                    bool                           `xml:"forceGenSeof"`
	EnableOOP                       bool                           `xml:"enableOOP"`
	CharstringCompat                bool                           `xml:"charstringCompat"`
	CxxCompiler                     string                         `xml:"CxxCompiler"`
	OptimizationLevel               string                         `xml:"optimizationLevel"`
	OtherOptimizationFlags          string                         `xml:"otherOptimizationFlags"`
	ProfiledFileList                *ResourceType                  `xml:"profiledFileList"`
	SolarisSpecificLibraries        *SolarisSpecificLibraries      `xml:"SolarisSpecificLibraries"`
	Solaris8SpecificLibraries       *Solaris8SpecificLibraries     `xml:"Solaris8SpecificLibraries"`
	FreeBSDSpecificLibraries        *FreeBSDSpecificLibraries      `xml:"FreeBSDSpecificLibraries"`
	LinuxSpecificLibraries          *LinuxSpecificLibraries        `xml:"LinuxSpecificLibraries"`
	Win32SpecificLibraries          *Win32SpecificLibraries        `xml:"Win32SpecificLibraries"`
	AdditionalObjects               *AdditionalObjects             `xml:"additionalObjects"`
	LinkerLibraries                 *LinkerLibraries               `xml:"linkerLibraries"`
	LinkerLibrarySearchPath         *LinkerLibrarySearchPath       `xml:"linkerLibrarySearchPath"`
	DisablePredefinedExternalFolder bool                           `xml:"disablePredefinedExternalFolder"`
	UseGoldLinker                   bool                           `xml:"useGoldLinker"`
	FreeTextLinkerOptions           string                         `xml:"freeTextLinkerOptions"`
	BuildLevel                      string                         `xml:"buildLevel"`
	ProjectSpecificRulesGenerator   *ProjectSpecificRulesGenerator `xml:"ProjectSpecificRulesGenerator"`
}

type LocalBuildSettings struct {
	MakefileFlags    string `xml:"MakefileFlags"`
	MakefileScript   string `xml:"MakefileScript"`
	WorkingDirectory string `xml:"workingDirectory"`
}

type RemoteHost struct {
	Active  bool   `xml:"Active"`
	Name    string `xml:"Name"`
	Command string `xml:"Command"`
}

type RemoteBuildProperties struct {
	RemoteHost               []*RemoteHost `xml:"RemoteHost"`
	ParallelCommandExecution bool          `xml:"ParallelCommandExecution"`
}

type NamingCoventions struct {
	EnableProjectSpecificSettings string `xml:"enableProjectSpecificSettings"`
	TTCN3ModuleName               string `xml:"TTCN3ModuleName"`
	ASN1ModuleName                string `xml:"ASN1ModuleName"`
	Altstep                       string `xml:"altstep"`
	GlobalConstant                string `xml:"globalConstant"`
	ExternalConstant              string `xml:"externalConstant"`
	Function                      string `xml:"function"`
	ExternalFunction              string `xml:"externalFunction"`
	ModuleParameter               string `xml:"moduleParameter"`
	GlobalPort                    string `xml:"globalPort"`
	GlobalTemplate                string `xml:"globalTemplate"`
	Testcase                      string `xml:"testcase"`
	GlobalTimer                   string `xml:"globalTimer"`
	Type                          string `xml:"type"`
	Group                         string `xml:"group"`
	LocalConstant                 string `xml:"localConstant"`
	LocalVariable                 string `xml:"localVariable"`
	LocalTemplate                 string `xml:"localTemplate"`
	LocalVariableTemplate         string `xml:"localVariableTemplate"`
	LocalTimer                    string `xml:"localTimer"`
	FormalParameter               string `xml:"formalParameter"`
	ComponentConstant             string `xml:"componentConstant"`
	ComponentVariable             string `xml:"componentVariable"`
	ComponentTimer                string `xml:"componentTimer"`
}

type ConfigurationRequirements struct {
	ConfigurationRequirement []*ConfigurationRequirementType `xml:"configurationRequirement"`
}

type ProjectProperties struct {
	MakefileSettings          *MakefileSettings          `xml:"MakefileSettings"`
	LocalBuildSettings        *LocalBuildSettings        `xml:"LocalBuildSettings"`
	RemoteBuildProperties     *RemoteBuildProperties     `xml:"RemoteBuildProperties"`
	NamingCoventions          *NamingCoventions          `xml:"NamingCoventions"`
	ConfigurationRequirements *ConfigurationRequirements `xml:"ConfigurationRequirements"`
}

type FolderProperties struct {
	ExcludeFromBuild bool              `xml:"ExcludeFromBuild"`
	CentralStorage   bool              `xml:"centralStorage"`
	NamingCoventions *NamingCoventions `xml:"NamingCoventions"`
}

type FolderResource struct {
	FolderPath       string            `xml:"FolderPath"`
	FolderProperties *FolderProperties `xml:"FolderProperties"`
}

type FileProperties struct {
	ExcludeFromBuild bool `xml:"ExcludeFromBuild"`
}

type FileResource struct {
	FilePath       string          `xml:"FilePath"`
	FileProperties *FileProperties `xml:"FileProperties"`
}

type ConfigurationType struct {
	ProjectProperties *ProjectProperties `xml:"ProjectProperties"`
	FolderProperties  *FolderProperties  `xml:"FolderProperties"`
	FileProperties    *FileProperties    `xml:"FileProperties"`
}

type NamedConfigurationType struct {
	NameAttr string `xml:"name,attr"`
	*ConfigurationType
}

type ConfigurationRequirementType struct {
	ProjectName            string `xml:"projectName"`
	RequiredConfiguration  string `xml:"requiredConfiguration"`
	RerquiredConfiguration string `xml:"rerquiredConfiguration"`
}

type ResourceType struct {
	ProjectRelativePathAttr string `xml:"projectRelativePath,attr"`
	RelativeURIAttr         string `xml:"relativeURI,attr,omitempty"`
	RawURIAttr              string `xml:"rawURI,attr,omitempty"`
}

type ReferencedProject struct {
	NameAttr               string `xml:"name,attr"`
	ProjectLocationURIAttr string `xml:"projectLocationURI,attr,omitempty"`
	TpdNameAttr            string `xml:"tpdName,attr,omitempty"`
}

type ReferencedProjects struct {
	ReferencedProject []*ReferencedProject `xml:"ReferencedProject"`
}

type Folders struct {
	FolderResource []*ResourceType `xml:"FolderResource"`
}

type Files struct {
	FileResource []*ResourceType `xml:"FileResource"`
}

type PathVariable struct {
	NameAttr  string `xml:"name,attr"`
	ValueAttr string `xml:"value,attr"`
}

type PathVariables struct {
	PathVariable []*PathVariable `xml:"PathVariable"`
}

type Configurations struct {
	Configuration []*NamedConfigurationType `xml:"Configuration"`
}

type ProjectType struct {
	ProjectName         string              `xml:"ProjectName"`
	ReferencedProjects  *ReferencedProjects `xml:"ReferencedProjects"`
	Folders             *Folders            `xml:"Folders"`
	Files               *Files              `xml:"Files"`
	PathVariables       *PathVariables      `xml:"PathVariables"`
	ActiveConfiguration string              `xml:"ActiveConfiguration"`
	Configurations      *Configurations     `xml:"Configurations"`
}

type PackedReferencedProjectsType struct {
	PackedReferencedProject []*ProjectType `xml:"PackedReferencedProject"`
}

type TopLevelProjectType struct {
	VersionAttr              float64                       `xml:"version,attr"`
	PackedReferencedProjects *PackedReferencedProjectsType `xml:"PackedReferencedProjects"`
	*ProjectType
}

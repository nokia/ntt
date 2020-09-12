package main

/*
#cgo LDFLAGS: -Wl,--unresolved-symbols=ignore-all

typedef void* value_t;
typedef void* type_t;

enum direction {
	IN    = 0,
	INOUT = 1,
	OUT   = 2,
};

enum verdict {
	NONE   = 0,
	PASS   = 1,
	INCONC = 2,
	FAIL   = 3,
	ERROR  = 4,
};

struct qualified_name {
	char *module;
	char *name;
	void *aux;
};

struct component {
	struct {
		char *data;
		int bits;
		void *aux;
	} instance;
	char *name;
	struct qualified_name type;
};

struct port {
	struct component component;
	char *name;
	int index;
	struct qualified_name type;
	void *aux;
};

struct parameter {
	char *name;
	enum direction direction;
	value_t value;
};

struct parameter_type {
	char *name;
	type_t type;
	enum direction direction;
};

struct module_parameter {
	struct qualified_name name;
	value_t value;
};


struct name_list {
	int len;
	struct qualified_name *data;
};

struct port_list {
	int len;
	struct port *data;
};

struct module_parameter_list {
	int len;
	struct module_parameter *data;
};

struct parameter_list {
	int len;
	struct parameter *data;
};

struct parameter_type_list {
	int len;
	struct parameter_type *data;
};


// Callbacks to be implemented by plugin.

void tciTestCaseStarted    (struct qualified_name id, struct parameter_list params, double timeout);
void tciTestCaseTerminated (value_t verdict, struct parameter_list params);
void tciControlTerminated  ();
void tciLog                (char* msg);
void tciError              (char* msg);
void tcinonMain();

value_t tciGetModulePar(struct qualified_name id);


// Functions provided by k3r.

void                         tciRootModule            (char* id);

struct name_list             tciGetImportedModules    ();
struct module_parameter_list tciGetModuleParameters   (struct qualified_name id);
struct name_list             tciGetTestCases          ();
struct parameter_type_list   tciGetTestCaseParameters (struct qualified_name id);
struct port_list             tciGetTestCaseTSI        (struct qualified_name id);
struct name_list             tcinonGetModulesList     ();
void                         tcinonK3rExit            ();

void             tciStartTestCase (struct qualified_name id, struct parameter_list params);
void             tciStopTestCase  ();
struct component tciStartControl  ();
void             tciStopControl   ();

int tciGetVerdictValue(value_t value);

value_t tciParseValue(type_t type, char* s);
void tcinonReleaseValue(value_t value);
*/
import "C"
import (
	"context"
	"log"
	"net"
	"strings"

	pb "github.com/nokia/ntt/protobuf"
	"google.golang.org/grpc"
)

//export tciTestCaseStarted
func tciTestCaseStarted(id C.struct_qualified_name, params C.struct_parameter_list, timeout float64) {
}

//export tciTestCaseTerminated
func tciTestCaseTerminated(verdict C.value_t, params C.struct_parameter_list) {}

//export tciControlTerminated
func tciControlTerminated() {}

//export tciLog
func tciLog(msg *C.char) {}

//export tciError
func tciError(msg *C.char) {}

//export tcinonMain
func tcinonMain() {
	main()
}

//export tciGetModulePar
func tciGetModulePar(id C.struct_qualified_name) C.value_t { return nil }

func str(id C.struct_qualified_name) string {
	var l []string
	if id.module != nil {
		l = append(l, C.GoString(id.module))
	}
	if id.name != nil {
		l = append(l, C.GoString(id.name))
	}
	return strings.Join(l, ".")
}

type server struct {
}

func (s *server) Run(ctx context.Context, in *pb.RunRequest) (*pb.RunResponse, error) {
	return nil, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50123")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterRuntimeServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	//C.tciRootModule(C.CString("main"))
	//tcs := C.tciGetTestCases()
	//l := int(tcs.len)
	//if l <= 0 {
	//	return
	//}
	//slice := (*[1 << 28]C.struct_qualified_name)(unsafe.Pointer(tcs.data))[:l:l]
	//for _, tc := range slice {
	//	fmt.Println(str(tc))
	//}
}

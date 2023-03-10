#include <k3/Plugin.hh>
namespace k3 {
using namespace v2;
}

void hello(const std::vector<k3::Value> &args, k3::Value ret) {
    k3::pllg("Hello ", k3::to_string(args[0]));
}

static struct Plugin {
    Plugin() {
	    registerPlugin("test.hello", hello);
    }
} p;

// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: value.proto

#include "value.pb.h"

#include <algorithm>

#include <google/protobuf/io/coded_stream.h>
#include <google/protobuf/extension_set.h>
#include <google/protobuf/wire_format_lite.h>
#include <google/protobuf/descriptor.h>
#include <google/protobuf/generated_message_reflection.h>
#include <google/protobuf/reflection_ops.h>
#include <google/protobuf/wire_format.h>
// @@protoc_insertion_point(includes)
#include <google/protobuf/port_def.inc>
extern PROTOBUF_INTERNAL_EXPORT_value_2eproto ::PROTOBUF_NAMESPACE_ID::internal::SCCInfo<0> scc_info_Composite_value_2eproto;
namespace ntt {
class ValueDefaultTypeInternal {
 public:
  ::PROTOBUF_NAMESPACE_ID::internal::ExplicitlyConstructed<Value> _instance;
  ::PROTOBUF_NAMESPACE_ID::internal::ArenaStringPtr byte_value_;
  bool bool_value_;
  int verdict_value_;
  ::PROTOBUF_NAMESPACE_ID::internal::ArenaStringPtr string_value_;
  double float_value_;
  ::PROTOBUF_NAMESPACE_ID::int32 int_value_;
  ::PROTOBUF_NAMESPACE_ID::internal::ArenaStringPtr big_value_;
  const ::ntt::Composite* composite_value_;
} _Value_default_instance_;
class CompositeDefaultTypeInternal {
 public:
  ::PROTOBUF_NAMESPACE_ID::internal::ExplicitlyConstructed<Composite> _instance;
} _Composite_default_instance_;
}  // namespace ntt
static void InitDefaultsscc_info_Composite_value_2eproto() {
  GOOGLE_PROTOBUF_VERIFY_VERSION;

  {
    void* ptr = &::ntt::_Value_default_instance_;
    new (ptr) ::ntt::Value();
    ::PROTOBUF_NAMESPACE_ID::internal::OnShutdownDestroyMessage(ptr);
  }
  {
    void* ptr = &::ntt::_Composite_default_instance_;
    new (ptr) ::ntt::Composite();
    ::PROTOBUF_NAMESPACE_ID::internal::OnShutdownDestroyMessage(ptr);
  }
  ::ntt::Value::InitAsDefaultInstance();
  ::ntt::Composite::InitAsDefaultInstance();
}

::PROTOBUF_NAMESPACE_ID::internal::SCCInfo<0> scc_info_Composite_value_2eproto =
    {{ATOMIC_VAR_INIT(::PROTOBUF_NAMESPACE_ID::internal::SCCInfoBase::kUninitialized), 0, 0, InitDefaultsscc_info_Composite_value_2eproto}, {}};

static ::PROTOBUF_NAMESPACE_ID::Metadata file_level_metadata_value_2eproto[2];
static const ::PROTOBUF_NAMESPACE_ID::EnumDescriptor* file_level_enum_descriptors_value_2eproto[1];
static constexpr ::PROTOBUF_NAMESPACE_ID::ServiceDescriptor const** file_level_service_descriptors_value_2eproto = nullptr;

const ::PROTOBUF_NAMESPACE_ID::uint32 TableStruct_value_2eproto::offsets[] PROTOBUF_SECTION_VARIABLE(protodesc_cold) = {
  ~0u,  // no _has_bits_
  PROTOBUF_FIELD_OFFSET(::ntt::Value, _internal_metadata_),
  ~0u,  // no _extensions_
  PROTOBUF_FIELD_OFFSET(::ntt::Value, _oneof_case_[0]),
  ~0u,  // no _weak_field_map_
  offsetof(::ntt::ValueDefaultTypeInternal, byte_value_),
  offsetof(::ntt::ValueDefaultTypeInternal, bool_value_),
  offsetof(::ntt::ValueDefaultTypeInternal, verdict_value_),
  offsetof(::ntt::ValueDefaultTypeInternal, string_value_),
  offsetof(::ntt::ValueDefaultTypeInternal, float_value_),
  offsetof(::ntt::ValueDefaultTypeInternal, int_value_),
  offsetof(::ntt::ValueDefaultTypeInternal, big_value_),
  offsetof(::ntt::ValueDefaultTypeInternal, composite_value_),
  PROTOBUF_FIELD_OFFSET(::ntt::Value, kind_),
  ~0u,  // no _has_bits_
  PROTOBUF_FIELD_OFFSET(::ntt::Composite, _internal_metadata_),
  ~0u,  // no _extensions_
  ~0u,  // no _oneof_case_
  ~0u,  // no _weak_field_map_
  PROTOBUF_FIELD_OFFSET(::ntt::Composite, values_),
};
static const ::PROTOBUF_NAMESPACE_ID::internal::MigrationSchema schemas[] PROTOBUF_SECTION_VARIABLE(protodesc_cold) = {
  { 0, -1, sizeof(::ntt::Value)},
  { 14, -1, sizeof(::ntt::Composite)},
};

static ::PROTOBUF_NAMESPACE_ID::Message const * const file_default_instances[] = {
  reinterpret_cast<const ::PROTOBUF_NAMESPACE_ID::Message*>(&::ntt::_Value_default_instance_),
  reinterpret_cast<const ::PROTOBUF_NAMESPACE_ID::Message*>(&::ntt::_Composite_default_instance_),
};

const char descriptor_table_protodef_value_2eproto[] PROTOBUF_SECTION_VARIABLE(protodesc_cold) =
  "\n\013value.proto\022\003ntt\"\346\001\n\005Value\022\024\n\nbyte_val"
  "ue\030\001 \001(\014H\000\022\024\n\nbool_value\030\002 \001(\010H\000\022%\n\rverd"
  "ict_value\030\003 \001(\0162\014.ntt.VerdictH\000\022\026\n\014strin"
  "g_value\030\004 \001(\tH\000\022\025\n\013float_value\030\005 \001(\001H\000\022\023"
  "\n\tint_value\030\006 \001(\005H\000\022\023\n\tbig_value\030\007 \001(\tH\000"
  "\022)\n\017composite_value\030\010 \001(\0132\016.ntt.Composit"
  "eH\000B\006\n\004kind\"\'\n\tComposite\022\032\n\006values\030\001 \003(\013"
  "2\n.ntt.Value*>\n\007Verdict\022\010\n\004NONE\020\000\022\010\n\004PAS"
  "S\020\001\022\n\n\006INCONC\020\002\022\010\n\004FAIL\020\003\022\t\n\005ERROR\020\004B\037Z\035"
  "github.com/nokia/ntt/protobufb\006proto3"
  ;
static const ::PROTOBUF_NAMESPACE_ID::internal::DescriptorTable*const descriptor_table_value_2eproto_deps[1] = {
};
static ::PROTOBUF_NAMESPACE_ID::internal::SCCInfoBase*const descriptor_table_value_2eproto_sccs[1] = {
  &scc_info_Composite_value_2eproto.base,
};
static ::PROTOBUF_NAMESPACE_ID::internal::once_flag descriptor_table_value_2eproto_once;
static bool descriptor_table_value_2eproto_initialized = false;
const ::PROTOBUF_NAMESPACE_ID::internal::DescriptorTable descriptor_table_value_2eproto = {
  &descriptor_table_value_2eproto_initialized, descriptor_table_protodef_value_2eproto, "value.proto", 397,
  &descriptor_table_value_2eproto_once, descriptor_table_value_2eproto_sccs, descriptor_table_value_2eproto_deps, 1, 0,
  schemas, file_default_instances, TableStruct_value_2eproto::offsets,
  file_level_metadata_value_2eproto, 2, file_level_enum_descriptors_value_2eproto, file_level_service_descriptors_value_2eproto,
};

// Force running AddDescriptors() at dynamic initialization time.
static bool dynamic_init_dummy_value_2eproto = (  ::PROTOBUF_NAMESPACE_ID::internal::AddDescriptors(&descriptor_table_value_2eproto), true);
namespace ntt {
const ::PROTOBUF_NAMESPACE_ID::EnumDescriptor* Verdict_descriptor() {
  ::PROTOBUF_NAMESPACE_ID::internal::AssignDescriptors(&descriptor_table_value_2eproto);
  return file_level_enum_descriptors_value_2eproto[0];
}
bool Verdict_IsValid(int value) {
  switch (value) {
    case 0:
    case 1:
    case 2:
    case 3:
    case 4:
      return true;
    default:
      return false;
  }
}


// ===================================================================

void Value::InitAsDefaultInstance() {
  ::ntt::_Value_default_instance_.byte_value_.UnsafeSetDefault(
      &::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  ::ntt::_Value_default_instance_.bool_value_ = false;
  ::ntt::_Value_default_instance_.verdict_value_ = 0;
  ::ntt::_Value_default_instance_.string_value_.UnsafeSetDefault(
      &::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  ::ntt::_Value_default_instance_.float_value_ = 0;
  ::ntt::_Value_default_instance_.int_value_ = 0;
  ::ntt::_Value_default_instance_.big_value_.UnsafeSetDefault(
      &::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  ::ntt::_Value_default_instance_.composite_value_ = const_cast< ::ntt::Composite*>(
      ::ntt::Composite::internal_default_instance());
}
class Value::_Internal {
 public:
  static const ::ntt::Composite& composite_value(const Value* msg);
};

const ::ntt::Composite&
Value::_Internal::composite_value(const Value* msg) {
  return *msg->kind_.composite_value_;
}
void Value::set_allocated_composite_value(::ntt::Composite* composite_value) {
  ::PROTOBUF_NAMESPACE_ID::Arena* message_arena = GetArenaNoVirtual();
  clear_kind();
  if (composite_value) {
    ::PROTOBUF_NAMESPACE_ID::Arena* submessage_arena = nullptr;
    if (message_arena != submessage_arena) {
      composite_value = ::PROTOBUF_NAMESPACE_ID::internal::GetOwnedMessage(
          message_arena, composite_value, submessage_arena);
    }
    set_has_composite_value();
    kind_.composite_value_ = composite_value;
  }
  // @@protoc_insertion_point(field_set_allocated:ntt.Value.composite_value)
}
Value::Value()
  : ::PROTOBUF_NAMESPACE_ID::Message(), _internal_metadata_(nullptr) {
  SharedCtor();
  // @@protoc_insertion_point(constructor:ntt.Value)
}
Value::Value(const Value& from)
  : ::PROTOBUF_NAMESPACE_ID::Message(),
      _internal_metadata_(nullptr) {
  _internal_metadata_.MergeFrom(from._internal_metadata_);
  clear_has_kind();
  switch (from.kind_case()) {
    case kByteValue: {
      _internal_set_byte_value(from._internal_byte_value());
      break;
    }
    case kBoolValue: {
      _internal_set_bool_value(from._internal_bool_value());
      break;
    }
    case kVerdictValue: {
      _internal_set_verdict_value(from._internal_verdict_value());
      break;
    }
    case kStringValue: {
      _internal_set_string_value(from._internal_string_value());
      break;
    }
    case kFloatValue: {
      _internal_set_float_value(from._internal_float_value());
      break;
    }
    case kIntValue: {
      _internal_set_int_value(from._internal_int_value());
      break;
    }
    case kBigValue: {
      _internal_set_big_value(from._internal_big_value());
      break;
    }
    case kCompositeValue: {
      _internal_mutable_composite_value()->::ntt::Composite::MergeFrom(from._internal_composite_value());
      break;
    }
    case KIND_NOT_SET: {
      break;
    }
  }
  // @@protoc_insertion_point(copy_constructor:ntt.Value)
}

void Value::SharedCtor() {
  ::PROTOBUF_NAMESPACE_ID::internal::InitSCC(&scc_info_Composite_value_2eproto.base);
  clear_has_kind();
}

Value::~Value() {
  // @@protoc_insertion_point(destructor:ntt.Value)
  SharedDtor();
}

void Value::SharedDtor() {
  if (has_kind()) {
    clear_kind();
  }
}

void Value::SetCachedSize(int size) const {
  _cached_size_.Set(size);
}
const Value& Value::default_instance() {
  ::PROTOBUF_NAMESPACE_ID::internal::InitSCC(&::scc_info_Composite_value_2eproto.base);
  return *internal_default_instance();
}


void Value::clear_kind() {
// @@protoc_insertion_point(one_of_clear_start:ntt.Value)
  switch (kind_case()) {
    case kByteValue: {
      kind_.byte_value_.DestroyNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
      break;
    }
    case kBoolValue: {
      // No need to clear
      break;
    }
    case kVerdictValue: {
      // No need to clear
      break;
    }
    case kStringValue: {
      kind_.string_value_.DestroyNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
      break;
    }
    case kFloatValue: {
      // No need to clear
      break;
    }
    case kIntValue: {
      // No need to clear
      break;
    }
    case kBigValue: {
      kind_.big_value_.DestroyNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
      break;
    }
    case kCompositeValue: {
      delete kind_.composite_value_;
      break;
    }
    case KIND_NOT_SET: {
      break;
    }
  }
  _oneof_case_[0] = KIND_NOT_SET;
}


void Value::Clear() {
// @@protoc_insertion_point(message_clear_start:ntt.Value)
  ::PROTOBUF_NAMESPACE_ID::uint32 cached_has_bits = 0;
  // Prevent compiler warnings about cached_has_bits being unused
  (void) cached_has_bits;

  clear_kind();
  _internal_metadata_.Clear();
}

const char* Value::_InternalParse(const char* ptr, ::PROTOBUF_NAMESPACE_ID::internal::ParseContext* ctx) {
#define CHK_(x) if (PROTOBUF_PREDICT_FALSE(!(x))) goto failure
  while (!ctx->Done(&ptr)) {
    ::PROTOBUF_NAMESPACE_ID::uint32 tag;
    ptr = ::PROTOBUF_NAMESPACE_ID::internal::ReadTag(ptr, &tag);
    CHK_(ptr);
    switch (tag >> 3) {
      // bytes byte_value = 1;
      case 1:
        if (PROTOBUF_PREDICT_TRUE(static_cast<::PROTOBUF_NAMESPACE_ID::uint8>(tag) == 10)) {
          auto str = _internal_mutable_byte_value();
          ptr = ::PROTOBUF_NAMESPACE_ID::internal::InlineGreedyStringParser(str, ptr, ctx);
          CHK_(ptr);
        } else goto handle_unusual;
        continue;
      // bool bool_value = 2;
      case 2:
        if (PROTOBUF_PREDICT_TRUE(static_cast<::PROTOBUF_NAMESPACE_ID::uint8>(tag) == 16)) {
          _internal_set_bool_value(::PROTOBUF_NAMESPACE_ID::internal::ReadVarint(&ptr));
          CHK_(ptr);
        } else goto handle_unusual;
        continue;
      // .ntt.Verdict verdict_value = 3;
      case 3:
        if (PROTOBUF_PREDICT_TRUE(static_cast<::PROTOBUF_NAMESPACE_ID::uint8>(tag) == 24)) {
          ::PROTOBUF_NAMESPACE_ID::uint64 val = ::PROTOBUF_NAMESPACE_ID::internal::ReadVarint(&ptr);
          CHK_(ptr);
          _internal_set_verdict_value(static_cast<::ntt::Verdict>(val));
        } else goto handle_unusual;
        continue;
      // string string_value = 4;
      case 4:
        if (PROTOBUF_PREDICT_TRUE(static_cast<::PROTOBUF_NAMESPACE_ID::uint8>(tag) == 34)) {
          auto str = _internal_mutable_string_value();
          ptr = ::PROTOBUF_NAMESPACE_ID::internal::InlineGreedyStringParser(str, ptr, ctx);
          CHK_(::PROTOBUF_NAMESPACE_ID::internal::VerifyUTF8(str, "ntt.Value.string_value"));
          CHK_(ptr);
        } else goto handle_unusual;
        continue;
      // double float_value = 5;
      case 5:
        if (PROTOBUF_PREDICT_TRUE(static_cast<::PROTOBUF_NAMESPACE_ID::uint8>(tag) == 41)) {
          _internal_set_float_value(::PROTOBUF_NAMESPACE_ID::internal::UnalignedLoad<double>(ptr));
          ptr += sizeof(double);
        } else goto handle_unusual;
        continue;
      // int32 int_value = 6;
      case 6:
        if (PROTOBUF_PREDICT_TRUE(static_cast<::PROTOBUF_NAMESPACE_ID::uint8>(tag) == 48)) {
          _internal_set_int_value(::PROTOBUF_NAMESPACE_ID::internal::ReadVarint(&ptr));
          CHK_(ptr);
        } else goto handle_unusual;
        continue;
      // string big_value = 7;
      case 7:
        if (PROTOBUF_PREDICT_TRUE(static_cast<::PROTOBUF_NAMESPACE_ID::uint8>(tag) == 58)) {
          auto str = _internal_mutable_big_value();
          ptr = ::PROTOBUF_NAMESPACE_ID::internal::InlineGreedyStringParser(str, ptr, ctx);
          CHK_(::PROTOBUF_NAMESPACE_ID::internal::VerifyUTF8(str, "ntt.Value.big_value"));
          CHK_(ptr);
        } else goto handle_unusual;
        continue;
      // .ntt.Composite composite_value = 8;
      case 8:
        if (PROTOBUF_PREDICT_TRUE(static_cast<::PROTOBUF_NAMESPACE_ID::uint8>(tag) == 66)) {
          ptr = ctx->ParseMessage(_internal_mutable_composite_value(), ptr);
          CHK_(ptr);
        } else goto handle_unusual;
        continue;
      default: {
      handle_unusual:
        if ((tag & 7) == 4 || tag == 0) {
          ctx->SetLastTag(tag);
          goto success;
        }
        ptr = UnknownFieldParse(tag, &_internal_metadata_, ptr, ctx);
        CHK_(ptr != nullptr);
        continue;
      }
    }  // switch
  }  // while
success:
  return ptr;
failure:
  ptr = nullptr;
  goto success;
#undef CHK_
}

::PROTOBUF_NAMESPACE_ID::uint8* Value::_InternalSerialize(
    ::PROTOBUF_NAMESPACE_ID::uint8* target, ::PROTOBUF_NAMESPACE_ID::io::EpsCopyOutputStream* stream) const {
  // @@protoc_insertion_point(serialize_to_array_start:ntt.Value)
  ::PROTOBUF_NAMESPACE_ID::uint32 cached_has_bits = 0;
  (void) cached_has_bits;

  // bytes byte_value = 1;
  if (_internal_has_byte_value()) {
    target = stream->WriteBytesMaybeAliased(
        1, this->_internal_byte_value(), target);
  }

  // bool bool_value = 2;
  if (_internal_has_bool_value()) {
    target = stream->EnsureSpace(target);
    target = ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::WriteBoolToArray(2, this->_internal_bool_value(), target);
  }

  // .ntt.Verdict verdict_value = 3;
  if (_internal_has_verdict_value()) {
    target = stream->EnsureSpace(target);
    target = ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::WriteEnumToArray(
      3, this->_internal_verdict_value(), target);
  }

  // string string_value = 4;
  if (_internal_has_string_value()) {
    ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::VerifyUtf8String(
      this->_internal_string_value().data(), static_cast<int>(this->_internal_string_value().length()),
      ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::SERIALIZE,
      "ntt.Value.string_value");
    target = stream->WriteStringMaybeAliased(
        4, this->_internal_string_value(), target);
  }

  // double float_value = 5;
  if (_internal_has_float_value()) {
    target = stream->EnsureSpace(target);
    target = ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::WriteDoubleToArray(5, this->_internal_float_value(), target);
  }

  // int32 int_value = 6;
  if (_internal_has_int_value()) {
    target = stream->EnsureSpace(target);
    target = ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::WriteInt32ToArray(6, this->_internal_int_value(), target);
  }

  // string big_value = 7;
  if (_internal_has_big_value()) {
    ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::VerifyUtf8String(
      this->_internal_big_value().data(), static_cast<int>(this->_internal_big_value().length()),
      ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::SERIALIZE,
      "ntt.Value.big_value");
    target = stream->WriteStringMaybeAliased(
        7, this->_internal_big_value(), target);
  }

  // .ntt.Composite composite_value = 8;
  if (_internal_has_composite_value()) {
    target = stream->EnsureSpace(target);
    target = ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::
      InternalWriteMessage(
        8, _Internal::composite_value(this), target, stream);
  }

  if (PROTOBUF_PREDICT_FALSE(_internal_metadata_.have_unknown_fields())) {
    target = ::PROTOBUF_NAMESPACE_ID::internal::WireFormat::InternalSerializeUnknownFieldsToArray(
        _internal_metadata_.unknown_fields(), target, stream);
  }
  // @@protoc_insertion_point(serialize_to_array_end:ntt.Value)
  return target;
}

size_t Value::ByteSizeLong() const {
// @@protoc_insertion_point(message_byte_size_start:ntt.Value)
  size_t total_size = 0;

  ::PROTOBUF_NAMESPACE_ID::uint32 cached_has_bits = 0;
  // Prevent compiler warnings about cached_has_bits being unused
  (void) cached_has_bits;

  switch (kind_case()) {
    // bytes byte_value = 1;
    case kByteValue: {
      total_size += 1 +
        ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::BytesSize(
          this->_internal_byte_value());
      break;
    }
    // bool bool_value = 2;
    case kBoolValue: {
      total_size += 1 + 1;
      break;
    }
    // .ntt.Verdict verdict_value = 3;
    case kVerdictValue: {
      total_size += 1 +
        ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::EnumSize(this->_internal_verdict_value());
      break;
    }
    // string string_value = 4;
    case kStringValue: {
      total_size += 1 +
        ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::StringSize(
          this->_internal_string_value());
      break;
    }
    // double float_value = 5;
    case kFloatValue: {
      total_size += 1 + 8;
      break;
    }
    // int32 int_value = 6;
    case kIntValue: {
      total_size += 1 +
        ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::Int32Size(
          this->_internal_int_value());
      break;
    }
    // string big_value = 7;
    case kBigValue: {
      total_size += 1 +
        ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::StringSize(
          this->_internal_big_value());
      break;
    }
    // .ntt.Composite composite_value = 8;
    case kCompositeValue: {
      total_size += 1 +
        ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::MessageSize(
          *kind_.composite_value_);
      break;
    }
    case KIND_NOT_SET: {
      break;
    }
  }
  if (PROTOBUF_PREDICT_FALSE(_internal_metadata_.have_unknown_fields())) {
    return ::PROTOBUF_NAMESPACE_ID::internal::ComputeUnknownFieldsSize(
        _internal_metadata_, total_size, &_cached_size_);
  }
  int cached_size = ::PROTOBUF_NAMESPACE_ID::internal::ToCachedSize(total_size);
  SetCachedSize(cached_size);
  return total_size;
}

void Value::MergeFrom(const ::PROTOBUF_NAMESPACE_ID::Message& from) {
// @@protoc_insertion_point(generalized_merge_from_start:ntt.Value)
  GOOGLE_DCHECK_NE(&from, this);
  const Value* source =
      ::PROTOBUF_NAMESPACE_ID::DynamicCastToGenerated<Value>(
          &from);
  if (source == nullptr) {
  // @@protoc_insertion_point(generalized_merge_from_cast_fail:ntt.Value)
    ::PROTOBUF_NAMESPACE_ID::internal::ReflectionOps::Merge(from, this);
  } else {
  // @@protoc_insertion_point(generalized_merge_from_cast_success:ntt.Value)
    MergeFrom(*source);
  }
}

void Value::MergeFrom(const Value& from) {
// @@protoc_insertion_point(class_specific_merge_from_start:ntt.Value)
  GOOGLE_DCHECK_NE(&from, this);
  _internal_metadata_.MergeFrom(from._internal_metadata_);
  ::PROTOBUF_NAMESPACE_ID::uint32 cached_has_bits = 0;
  (void) cached_has_bits;

  switch (from.kind_case()) {
    case kByteValue: {
      _internal_set_byte_value(from._internal_byte_value());
      break;
    }
    case kBoolValue: {
      _internal_set_bool_value(from._internal_bool_value());
      break;
    }
    case kVerdictValue: {
      _internal_set_verdict_value(from._internal_verdict_value());
      break;
    }
    case kStringValue: {
      _internal_set_string_value(from._internal_string_value());
      break;
    }
    case kFloatValue: {
      _internal_set_float_value(from._internal_float_value());
      break;
    }
    case kIntValue: {
      _internal_set_int_value(from._internal_int_value());
      break;
    }
    case kBigValue: {
      _internal_set_big_value(from._internal_big_value());
      break;
    }
    case kCompositeValue: {
      _internal_mutable_composite_value()->::ntt::Composite::MergeFrom(from._internal_composite_value());
      break;
    }
    case KIND_NOT_SET: {
      break;
    }
  }
}

void Value::CopyFrom(const ::PROTOBUF_NAMESPACE_ID::Message& from) {
// @@protoc_insertion_point(generalized_copy_from_start:ntt.Value)
  if (&from == this) return;
  Clear();
  MergeFrom(from);
}

void Value::CopyFrom(const Value& from) {
// @@protoc_insertion_point(class_specific_copy_from_start:ntt.Value)
  if (&from == this) return;
  Clear();
  MergeFrom(from);
}

bool Value::IsInitialized() const {
  return true;
}

void Value::InternalSwap(Value* other) {
  using std::swap;
  _internal_metadata_.Swap(&other->_internal_metadata_);
  swap(kind_, other->kind_);
  swap(_oneof_case_[0], other->_oneof_case_[0]);
}

::PROTOBUF_NAMESPACE_ID::Metadata Value::GetMetadata() const {
  return GetMetadataStatic();
}


// ===================================================================

void Composite::InitAsDefaultInstance() {
}
class Composite::_Internal {
 public:
};

Composite::Composite()
  : ::PROTOBUF_NAMESPACE_ID::Message(), _internal_metadata_(nullptr) {
  SharedCtor();
  // @@protoc_insertion_point(constructor:ntt.Composite)
}
Composite::Composite(const Composite& from)
  : ::PROTOBUF_NAMESPACE_ID::Message(),
      _internal_metadata_(nullptr),
      values_(from.values_) {
  _internal_metadata_.MergeFrom(from._internal_metadata_);
  // @@protoc_insertion_point(copy_constructor:ntt.Composite)
}

void Composite::SharedCtor() {
  ::PROTOBUF_NAMESPACE_ID::internal::InitSCC(&scc_info_Composite_value_2eproto.base);
}

Composite::~Composite() {
  // @@protoc_insertion_point(destructor:ntt.Composite)
  SharedDtor();
}

void Composite::SharedDtor() {
}

void Composite::SetCachedSize(int size) const {
  _cached_size_.Set(size);
}
const Composite& Composite::default_instance() {
  ::PROTOBUF_NAMESPACE_ID::internal::InitSCC(&::scc_info_Composite_value_2eproto.base);
  return *internal_default_instance();
}


void Composite::Clear() {
// @@protoc_insertion_point(message_clear_start:ntt.Composite)
  ::PROTOBUF_NAMESPACE_ID::uint32 cached_has_bits = 0;
  // Prevent compiler warnings about cached_has_bits being unused
  (void) cached_has_bits;

  values_.Clear();
  _internal_metadata_.Clear();
}

const char* Composite::_InternalParse(const char* ptr, ::PROTOBUF_NAMESPACE_ID::internal::ParseContext* ctx) {
#define CHK_(x) if (PROTOBUF_PREDICT_FALSE(!(x))) goto failure
  while (!ctx->Done(&ptr)) {
    ::PROTOBUF_NAMESPACE_ID::uint32 tag;
    ptr = ::PROTOBUF_NAMESPACE_ID::internal::ReadTag(ptr, &tag);
    CHK_(ptr);
    switch (tag >> 3) {
      // repeated .ntt.Value values = 1;
      case 1:
        if (PROTOBUF_PREDICT_TRUE(static_cast<::PROTOBUF_NAMESPACE_ID::uint8>(tag) == 10)) {
          ptr -= 1;
          do {
            ptr += 1;
            ptr = ctx->ParseMessage(_internal_add_values(), ptr);
            CHK_(ptr);
            if (!ctx->DataAvailable(ptr)) break;
          } while (::PROTOBUF_NAMESPACE_ID::internal::ExpectTag<10>(ptr));
        } else goto handle_unusual;
        continue;
      default: {
      handle_unusual:
        if ((tag & 7) == 4 || tag == 0) {
          ctx->SetLastTag(tag);
          goto success;
        }
        ptr = UnknownFieldParse(tag, &_internal_metadata_, ptr, ctx);
        CHK_(ptr != nullptr);
        continue;
      }
    }  // switch
  }  // while
success:
  return ptr;
failure:
  ptr = nullptr;
  goto success;
#undef CHK_
}

::PROTOBUF_NAMESPACE_ID::uint8* Composite::_InternalSerialize(
    ::PROTOBUF_NAMESPACE_ID::uint8* target, ::PROTOBUF_NAMESPACE_ID::io::EpsCopyOutputStream* stream) const {
  // @@protoc_insertion_point(serialize_to_array_start:ntt.Composite)
  ::PROTOBUF_NAMESPACE_ID::uint32 cached_has_bits = 0;
  (void) cached_has_bits;

  // repeated .ntt.Value values = 1;
  for (unsigned int i = 0,
      n = static_cast<unsigned int>(this->_internal_values_size()); i < n; i++) {
    target = stream->EnsureSpace(target);
    target = ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::
      InternalWriteMessage(1, this->_internal_values(i), target, stream);
  }

  if (PROTOBUF_PREDICT_FALSE(_internal_metadata_.have_unknown_fields())) {
    target = ::PROTOBUF_NAMESPACE_ID::internal::WireFormat::InternalSerializeUnknownFieldsToArray(
        _internal_metadata_.unknown_fields(), target, stream);
  }
  // @@protoc_insertion_point(serialize_to_array_end:ntt.Composite)
  return target;
}

size_t Composite::ByteSizeLong() const {
// @@protoc_insertion_point(message_byte_size_start:ntt.Composite)
  size_t total_size = 0;

  ::PROTOBUF_NAMESPACE_ID::uint32 cached_has_bits = 0;
  // Prevent compiler warnings about cached_has_bits being unused
  (void) cached_has_bits;

  // repeated .ntt.Value values = 1;
  total_size += 1UL * this->_internal_values_size();
  for (const auto& msg : this->values_) {
    total_size +=
      ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::MessageSize(msg);
  }

  if (PROTOBUF_PREDICT_FALSE(_internal_metadata_.have_unknown_fields())) {
    return ::PROTOBUF_NAMESPACE_ID::internal::ComputeUnknownFieldsSize(
        _internal_metadata_, total_size, &_cached_size_);
  }
  int cached_size = ::PROTOBUF_NAMESPACE_ID::internal::ToCachedSize(total_size);
  SetCachedSize(cached_size);
  return total_size;
}

void Composite::MergeFrom(const ::PROTOBUF_NAMESPACE_ID::Message& from) {
// @@protoc_insertion_point(generalized_merge_from_start:ntt.Composite)
  GOOGLE_DCHECK_NE(&from, this);
  const Composite* source =
      ::PROTOBUF_NAMESPACE_ID::DynamicCastToGenerated<Composite>(
          &from);
  if (source == nullptr) {
  // @@protoc_insertion_point(generalized_merge_from_cast_fail:ntt.Composite)
    ::PROTOBUF_NAMESPACE_ID::internal::ReflectionOps::Merge(from, this);
  } else {
  // @@protoc_insertion_point(generalized_merge_from_cast_success:ntt.Composite)
    MergeFrom(*source);
  }
}

void Composite::MergeFrom(const Composite& from) {
// @@protoc_insertion_point(class_specific_merge_from_start:ntt.Composite)
  GOOGLE_DCHECK_NE(&from, this);
  _internal_metadata_.MergeFrom(from._internal_metadata_);
  ::PROTOBUF_NAMESPACE_ID::uint32 cached_has_bits = 0;
  (void) cached_has_bits;

  values_.MergeFrom(from.values_);
}

void Composite::CopyFrom(const ::PROTOBUF_NAMESPACE_ID::Message& from) {
// @@protoc_insertion_point(generalized_copy_from_start:ntt.Composite)
  if (&from == this) return;
  Clear();
  MergeFrom(from);
}

void Composite::CopyFrom(const Composite& from) {
// @@protoc_insertion_point(class_specific_copy_from_start:ntt.Composite)
  if (&from == this) return;
  Clear();
  MergeFrom(from);
}

bool Composite::IsInitialized() const {
  return true;
}

void Composite::InternalSwap(Composite* other) {
  using std::swap;
  _internal_metadata_.Swap(&other->_internal_metadata_);
  values_.InternalSwap(&other->values_);
}

::PROTOBUF_NAMESPACE_ID::Metadata Composite::GetMetadata() const {
  return GetMetadataStatic();
}


// @@protoc_insertion_point(namespace_scope)
}  // namespace ntt
PROTOBUF_NAMESPACE_OPEN
template<> PROTOBUF_NOINLINE ::ntt::Value* Arena::CreateMaybeMessage< ::ntt::Value >(Arena* arena) {
  return Arena::CreateInternal< ::ntt::Value >(arena);
}
template<> PROTOBUF_NOINLINE ::ntt::Composite* Arena::CreateMaybeMessage< ::ntt::Composite >(Arena* arena) {
  return Arena::CreateInternal< ::ntt::Composite >(arena);
}
PROTOBUF_NAMESPACE_CLOSE

// @@protoc_insertion_point(global_scope)
#include <google/protobuf/port_undef.inc>
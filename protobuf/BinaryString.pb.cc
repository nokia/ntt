// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: BinaryString.proto

#include "BinaryString.pb.h"

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
namespace ntt {
class BinaryStringDefaultTypeInternal {
 public:
  ::PROTOBUF_NAMESPACE_ID::internal::ExplicitlyConstructed<BinaryString> _instance;
} _BinaryString_default_instance_;
}  // namespace ntt
static void InitDefaultsscc_info_BinaryString_BinaryString_2eproto() {
  GOOGLE_PROTOBUF_VERIFY_VERSION;

  {
    void* ptr = &::ntt::_BinaryString_default_instance_;
    new (ptr) ::ntt::BinaryString();
    ::PROTOBUF_NAMESPACE_ID::internal::OnShutdownDestroyMessage(ptr);
  }
  ::ntt::BinaryString::InitAsDefaultInstance();
}

::PROTOBUF_NAMESPACE_ID::internal::SCCInfo<0> scc_info_BinaryString_BinaryString_2eproto =
    {{ATOMIC_VAR_INIT(::PROTOBUF_NAMESPACE_ID::internal::SCCInfoBase::kUninitialized), 0, 0, InitDefaultsscc_info_BinaryString_BinaryString_2eproto}, {}};

static ::PROTOBUF_NAMESPACE_ID::Metadata file_level_metadata_BinaryString_2eproto[1];
static constexpr ::PROTOBUF_NAMESPACE_ID::EnumDescriptor const** file_level_enum_descriptors_BinaryString_2eproto = nullptr;
static constexpr ::PROTOBUF_NAMESPACE_ID::ServiceDescriptor const** file_level_service_descriptors_BinaryString_2eproto = nullptr;

const ::PROTOBUF_NAMESPACE_ID::uint32 TableStruct_BinaryString_2eproto::offsets[] PROTOBUF_SECTION_VARIABLE(protodesc_cold) = {
  ~0u,  // no _has_bits_
  PROTOBUF_FIELD_OFFSET(::ntt::BinaryString, _internal_metadata_),
  ~0u,  // no _extensions_
  ~0u,  // no _oneof_case_
  ~0u,  // no _weak_field_map_
  PROTOBUF_FIELD_OFFSET(::ntt::BinaryString, data_),
  PROTOBUF_FIELD_OFFSET(::ntt::BinaryString, nbits_),
};
static const ::PROTOBUF_NAMESPACE_ID::internal::MigrationSchema schemas[] PROTOBUF_SECTION_VARIABLE(protodesc_cold) = {
  { 0, -1, sizeof(::ntt::BinaryString)},
};

static ::PROTOBUF_NAMESPACE_ID::Message const * const file_default_instances[] = {
  reinterpret_cast<const ::PROTOBUF_NAMESPACE_ID::Message*>(&::ntt::_BinaryString_default_instance_),
};

const char descriptor_table_protodef_BinaryString_2eproto[] PROTOBUF_SECTION_VARIABLE(protodesc_cold) =
  "\n\022BinaryString.proto\022\003ntt\"+\n\014BinaryStrin"
  "g\022\014\n\004data\030\001 \001(\014\022\r\n\005nbits\030\002 \001(\005B\037Z\035github"
  ".com/nokia/ntt/protobufb\006proto3"
  ;
static const ::PROTOBUF_NAMESPACE_ID::internal::DescriptorTable*const descriptor_table_BinaryString_2eproto_deps[1] = {
};
static ::PROTOBUF_NAMESPACE_ID::internal::SCCInfoBase*const descriptor_table_BinaryString_2eproto_sccs[1] = {
  &scc_info_BinaryString_BinaryString_2eproto.base,
};
static ::PROTOBUF_NAMESPACE_ID::internal::once_flag descriptor_table_BinaryString_2eproto_once;
static bool descriptor_table_BinaryString_2eproto_initialized = false;
const ::PROTOBUF_NAMESPACE_ID::internal::DescriptorTable descriptor_table_BinaryString_2eproto = {
  &descriptor_table_BinaryString_2eproto_initialized, descriptor_table_protodef_BinaryString_2eproto, "BinaryString.proto", 111,
  &descriptor_table_BinaryString_2eproto_once, descriptor_table_BinaryString_2eproto_sccs, descriptor_table_BinaryString_2eproto_deps, 1, 0,
  schemas, file_default_instances, TableStruct_BinaryString_2eproto::offsets,
  file_level_metadata_BinaryString_2eproto, 1, file_level_enum_descriptors_BinaryString_2eproto, file_level_service_descriptors_BinaryString_2eproto,
};

// Force running AddDescriptors() at dynamic initialization time.
static bool dynamic_init_dummy_BinaryString_2eproto = (  ::PROTOBUF_NAMESPACE_ID::internal::AddDescriptors(&descriptor_table_BinaryString_2eproto), true);
namespace ntt {

// ===================================================================

void BinaryString::InitAsDefaultInstance() {
}
class BinaryString::_Internal {
 public:
};

BinaryString::BinaryString()
  : ::PROTOBUF_NAMESPACE_ID::Message(), _internal_metadata_(nullptr) {
  SharedCtor();
  // @@protoc_insertion_point(constructor:ntt.BinaryString)
}
BinaryString::BinaryString(const BinaryString& from)
  : ::PROTOBUF_NAMESPACE_ID::Message(),
      _internal_metadata_(nullptr) {
  _internal_metadata_.MergeFrom(from._internal_metadata_);
  data_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  if (!from._internal_data().empty()) {
    data_.AssignWithDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited(), from.data_);
  }
  nbits_ = from.nbits_;
  // @@protoc_insertion_point(copy_constructor:ntt.BinaryString)
}

void BinaryString::SharedCtor() {
  ::PROTOBUF_NAMESPACE_ID::internal::InitSCC(&scc_info_BinaryString_BinaryString_2eproto.base);
  data_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  nbits_ = 0;
}

BinaryString::~BinaryString() {
  // @@protoc_insertion_point(destructor:ntt.BinaryString)
  SharedDtor();
}

void BinaryString::SharedDtor() {
  data_.DestroyNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
}

void BinaryString::SetCachedSize(int size) const {
  _cached_size_.Set(size);
}
const BinaryString& BinaryString::default_instance() {
  ::PROTOBUF_NAMESPACE_ID::internal::InitSCC(&::scc_info_BinaryString_BinaryString_2eproto.base);
  return *internal_default_instance();
}


void BinaryString::Clear() {
// @@protoc_insertion_point(message_clear_start:ntt.BinaryString)
  ::PROTOBUF_NAMESPACE_ID::uint32 cached_has_bits = 0;
  // Prevent compiler warnings about cached_has_bits being unused
  (void) cached_has_bits;

  data_.ClearToEmptyNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  nbits_ = 0;
  _internal_metadata_.Clear();
}

const char* BinaryString::_InternalParse(const char* ptr, ::PROTOBUF_NAMESPACE_ID::internal::ParseContext* ctx) {
#define CHK_(x) if (PROTOBUF_PREDICT_FALSE(!(x))) goto failure
  while (!ctx->Done(&ptr)) {
    ::PROTOBUF_NAMESPACE_ID::uint32 tag;
    ptr = ::PROTOBUF_NAMESPACE_ID::internal::ReadTag(ptr, &tag);
    CHK_(ptr);
    switch (tag >> 3) {
      // bytes data = 1;
      case 1:
        if (PROTOBUF_PREDICT_TRUE(static_cast<::PROTOBUF_NAMESPACE_ID::uint8>(tag) == 10)) {
          auto str = _internal_mutable_data();
          ptr = ::PROTOBUF_NAMESPACE_ID::internal::InlineGreedyStringParser(str, ptr, ctx);
          CHK_(ptr);
        } else goto handle_unusual;
        continue;
      // int32 nbits = 2;
      case 2:
        if (PROTOBUF_PREDICT_TRUE(static_cast<::PROTOBUF_NAMESPACE_ID::uint8>(tag) == 16)) {
          nbits_ = ::PROTOBUF_NAMESPACE_ID::internal::ReadVarint(&ptr);
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

::PROTOBUF_NAMESPACE_ID::uint8* BinaryString::_InternalSerialize(
    ::PROTOBUF_NAMESPACE_ID::uint8* target, ::PROTOBUF_NAMESPACE_ID::io::EpsCopyOutputStream* stream) const {
  // @@protoc_insertion_point(serialize_to_array_start:ntt.BinaryString)
  ::PROTOBUF_NAMESPACE_ID::uint32 cached_has_bits = 0;
  (void) cached_has_bits;

  // bytes data = 1;
  if (this->data().size() > 0) {
    target = stream->WriteBytesMaybeAliased(
        1, this->_internal_data(), target);
  }

  // int32 nbits = 2;
  if (this->nbits() != 0) {
    target = stream->EnsureSpace(target);
    target = ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::WriteInt32ToArray(2, this->_internal_nbits(), target);
  }

  if (PROTOBUF_PREDICT_FALSE(_internal_metadata_.have_unknown_fields())) {
    target = ::PROTOBUF_NAMESPACE_ID::internal::WireFormat::InternalSerializeUnknownFieldsToArray(
        _internal_metadata_.unknown_fields(), target, stream);
  }
  // @@protoc_insertion_point(serialize_to_array_end:ntt.BinaryString)
  return target;
}

size_t BinaryString::ByteSizeLong() const {
// @@protoc_insertion_point(message_byte_size_start:ntt.BinaryString)
  size_t total_size = 0;

  ::PROTOBUF_NAMESPACE_ID::uint32 cached_has_bits = 0;
  // Prevent compiler warnings about cached_has_bits being unused
  (void) cached_has_bits;

  // bytes data = 1;
  if (this->data().size() > 0) {
    total_size += 1 +
      ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::BytesSize(
        this->_internal_data());
  }

  // int32 nbits = 2;
  if (this->nbits() != 0) {
    total_size += 1 +
      ::PROTOBUF_NAMESPACE_ID::internal::WireFormatLite::Int32Size(
        this->_internal_nbits());
  }

  if (PROTOBUF_PREDICT_FALSE(_internal_metadata_.have_unknown_fields())) {
    return ::PROTOBUF_NAMESPACE_ID::internal::ComputeUnknownFieldsSize(
        _internal_metadata_, total_size, &_cached_size_);
  }
  int cached_size = ::PROTOBUF_NAMESPACE_ID::internal::ToCachedSize(total_size);
  SetCachedSize(cached_size);
  return total_size;
}

void BinaryString::MergeFrom(const ::PROTOBUF_NAMESPACE_ID::Message& from) {
// @@protoc_insertion_point(generalized_merge_from_start:ntt.BinaryString)
  GOOGLE_DCHECK_NE(&from, this);
  const BinaryString* source =
      ::PROTOBUF_NAMESPACE_ID::DynamicCastToGenerated<BinaryString>(
          &from);
  if (source == nullptr) {
  // @@protoc_insertion_point(generalized_merge_from_cast_fail:ntt.BinaryString)
    ::PROTOBUF_NAMESPACE_ID::internal::ReflectionOps::Merge(from, this);
  } else {
  // @@protoc_insertion_point(generalized_merge_from_cast_success:ntt.BinaryString)
    MergeFrom(*source);
  }
}

void BinaryString::MergeFrom(const BinaryString& from) {
// @@protoc_insertion_point(class_specific_merge_from_start:ntt.BinaryString)
  GOOGLE_DCHECK_NE(&from, this);
  _internal_metadata_.MergeFrom(from._internal_metadata_);
  ::PROTOBUF_NAMESPACE_ID::uint32 cached_has_bits = 0;
  (void) cached_has_bits;

  if (from.data().size() > 0) {

    data_.AssignWithDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited(), from.data_);
  }
  if (from.nbits() != 0) {
    _internal_set_nbits(from._internal_nbits());
  }
}

void BinaryString::CopyFrom(const ::PROTOBUF_NAMESPACE_ID::Message& from) {
// @@protoc_insertion_point(generalized_copy_from_start:ntt.BinaryString)
  if (&from == this) return;
  Clear();
  MergeFrom(from);
}

void BinaryString::CopyFrom(const BinaryString& from) {
// @@protoc_insertion_point(class_specific_copy_from_start:ntt.BinaryString)
  if (&from == this) return;
  Clear();
  MergeFrom(from);
}

bool BinaryString::IsInitialized() const {
  return true;
}

void BinaryString::InternalSwap(BinaryString* other) {
  using std::swap;
  _internal_metadata_.Swap(&other->_internal_metadata_);
  data_.Swap(&other->data_, &::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited(),
    GetArenaNoVirtual());
  swap(nbits_, other->nbits_);
}

::PROTOBUF_NAMESPACE_ID::Metadata BinaryString::GetMetadata() const {
  return GetMetadataStatic();
}


// @@protoc_insertion_point(namespace_scope)
}  // namespace ntt
PROTOBUF_NAMESPACE_OPEN
template<> PROTOBUF_NOINLINE ::ntt::BinaryString* Arena::CreateMaybeMessage< ::ntt::BinaryString >(Arena* arena) {
  return Arena::CreateInternal< ::ntt::BinaryString >(arena);
}
PROTOBUF_NAMESPACE_CLOSE

// @@protoc_insertion_point(global_scope)
#include <google/protobuf/port_undef.inc>

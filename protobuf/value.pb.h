// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: value.proto

#ifndef GOOGLE_PROTOBUF_INCLUDED_value_2eproto
#define GOOGLE_PROTOBUF_INCLUDED_value_2eproto

#include <limits>
#include <string>

#include <google/protobuf/port_def.inc>
#if PROTOBUF_VERSION < 3011000
#error This file was generated by a newer version of protoc which is
#error incompatible with your Protocol Buffer headers. Please update
#error your headers.
#endif
#if 3011002 < PROTOBUF_MIN_PROTOC_VERSION
#error This file was generated by an older version of protoc which is
#error incompatible with your Protocol Buffer headers. Please
#error regenerate this file with a newer version of protoc.
#endif

#include <google/protobuf/port_undef.inc>
#include <google/protobuf/io/coded_stream.h>
#include <google/protobuf/arena.h>
#include <google/protobuf/arenastring.h>
#include <google/protobuf/generated_message_table_driven.h>
#include <google/protobuf/generated_message_util.h>
#include <google/protobuf/inlined_string_field.h>
#include <google/protobuf/metadata.h>
#include <google/protobuf/generated_message_reflection.h>
#include <google/protobuf/message.h>
#include <google/protobuf/repeated_field.h>  // IWYU pragma: export
#include <google/protobuf/extension_set.h>  // IWYU pragma: export
#include <google/protobuf/generated_enum_reflection.h>
#include <google/protobuf/unknown_field_set.h>
// @@protoc_insertion_point(includes)
#include <google/protobuf/port_def.inc>
#define PROTOBUF_INTERNAL_EXPORT_value_2eproto
PROTOBUF_NAMESPACE_OPEN
namespace internal {
class AnyMetadata;
}  // namespace internal
PROTOBUF_NAMESPACE_CLOSE

// Internal implementation detail -- do not use these members.
struct TableStruct_value_2eproto {
  static const ::PROTOBUF_NAMESPACE_ID::internal::ParseTableField entries[]
    PROTOBUF_SECTION_VARIABLE(protodesc_cold);
  static const ::PROTOBUF_NAMESPACE_ID::internal::AuxillaryParseTableField aux[]
    PROTOBUF_SECTION_VARIABLE(protodesc_cold);
  static const ::PROTOBUF_NAMESPACE_ID::internal::ParseTable schema[2]
    PROTOBUF_SECTION_VARIABLE(protodesc_cold);
  static const ::PROTOBUF_NAMESPACE_ID::internal::FieldMetadata field_metadata[];
  static const ::PROTOBUF_NAMESPACE_ID::internal::SerializationTable serialization_table[];
  static const ::PROTOBUF_NAMESPACE_ID::uint32 offsets[];
};
extern const ::PROTOBUF_NAMESPACE_ID::internal::DescriptorTable descriptor_table_value_2eproto;
namespace ntt {
class Composite;
class CompositeDefaultTypeInternal;
extern CompositeDefaultTypeInternal _Composite_default_instance_;
class Value;
class ValueDefaultTypeInternal;
extern ValueDefaultTypeInternal _Value_default_instance_;
}  // namespace ntt
PROTOBUF_NAMESPACE_OPEN
template<> ::ntt::Composite* Arena::CreateMaybeMessage<::ntt::Composite>(Arena*);
template<> ::ntt::Value* Arena::CreateMaybeMessage<::ntt::Value>(Arena*);
PROTOBUF_NAMESPACE_CLOSE
namespace ntt {

enum Verdict : int {
  NONE = 0,
  PASS = 1,
  INCONC = 2,
  FAIL = 3,
  ERROR = 4,
  Verdict_INT_MIN_SENTINEL_DO_NOT_USE_ = std::numeric_limits<::PROTOBUF_NAMESPACE_ID::int32>::min(),
  Verdict_INT_MAX_SENTINEL_DO_NOT_USE_ = std::numeric_limits<::PROTOBUF_NAMESPACE_ID::int32>::max()
};
bool Verdict_IsValid(int value);
constexpr Verdict Verdict_MIN = NONE;
constexpr Verdict Verdict_MAX = ERROR;
constexpr int Verdict_ARRAYSIZE = Verdict_MAX + 1;

const ::PROTOBUF_NAMESPACE_ID::EnumDescriptor* Verdict_descriptor();
template<typename T>
inline const std::string& Verdict_Name(T enum_t_value) {
  static_assert(::std::is_same<T, Verdict>::value ||
    ::std::is_integral<T>::value,
    "Incorrect type passed to function Verdict_Name.");
  return ::PROTOBUF_NAMESPACE_ID::internal::NameOfEnum(
    Verdict_descriptor(), enum_t_value);
}
inline bool Verdict_Parse(
    const std::string& name, Verdict* value) {
  return ::PROTOBUF_NAMESPACE_ID::internal::ParseNamedEnum<Verdict>(
    Verdict_descriptor(), name, value);
}
// ===================================================================

class Value :
    public ::PROTOBUF_NAMESPACE_ID::Message /* @@protoc_insertion_point(class_definition:ntt.Value) */ {
 public:
  Value();
  virtual ~Value();

  Value(const Value& from);
  Value(Value&& from) noexcept
    : Value() {
    *this = ::std::move(from);
  }

  inline Value& operator=(const Value& from) {
    CopyFrom(from);
    return *this;
  }
  inline Value& operator=(Value&& from) noexcept {
    if (GetArenaNoVirtual() == from.GetArenaNoVirtual()) {
      if (this != &from) InternalSwap(&from);
    } else {
      CopyFrom(from);
    }
    return *this;
  }

  static const ::PROTOBUF_NAMESPACE_ID::Descriptor* descriptor() {
    return GetDescriptor();
  }
  static const ::PROTOBUF_NAMESPACE_ID::Descriptor* GetDescriptor() {
    return GetMetadataStatic().descriptor;
  }
  static const ::PROTOBUF_NAMESPACE_ID::Reflection* GetReflection() {
    return GetMetadataStatic().reflection;
  }
  static const Value& default_instance();

  enum KindCase {
    kByteValue = 1,
    kBoolValue = 2,
    kVerdictValue = 3,
    kStringValue = 4,
    kFloatValue = 5,
    kIntValue = 6,
    kBigValue = 7,
    kCompositeValue = 8,
    KIND_NOT_SET = 0,
  };

  static void InitAsDefaultInstance();  // FOR INTERNAL USE ONLY
  static inline const Value* internal_default_instance() {
    return reinterpret_cast<const Value*>(
               &_Value_default_instance_);
  }
  static constexpr int kIndexInFileMessages =
    0;

  friend void swap(Value& a, Value& b) {
    a.Swap(&b);
  }
  inline void Swap(Value* other) {
    if (other == this) return;
    InternalSwap(other);
  }

  // implements Message ----------------------------------------------

  inline Value* New() const final {
    return CreateMaybeMessage<Value>(nullptr);
  }

  Value* New(::PROTOBUF_NAMESPACE_ID::Arena* arena) const final {
    return CreateMaybeMessage<Value>(arena);
  }
  void CopyFrom(const ::PROTOBUF_NAMESPACE_ID::Message& from) final;
  void MergeFrom(const ::PROTOBUF_NAMESPACE_ID::Message& from) final;
  void CopyFrom(const Value& from);
  void MergeFrom(const Value& from);
  PROTOBUF_ATTRIBUTE_REINITIALIZES void Clear() final;
  bool IsInitialized() const final;

  size_t ByteSizeLong() const final;
  const char* _InternalParse(const char* ptr, ::PROTOBUF_NAMESPACE_ID::internal::ParseContext* ctx) final;
  ::PROTOBUF_NAMESPACE_ID::uint8* _InternalSerialize(
      ::PROTOBUF_NAMESPACE_ID::uint8* target, ::PROTOBUF_NAMESPACE_ID::io::EpsCopyOutputStream* stream) const final;
  int GetCachedSize() const final { return _cached_size_.Get(); }

  private:
  inline void SharedCtor();
  inline void SharedDtor();
  void SetCachedSize(int size) const final;
  void InternalSwap(Value* other);
  friend class ::PROTOBUF_NAMESPACE_ID::internal::AnyMetadata;
  static ::PROTOBUF_NAMESPACE_ID::StringPiece FullMessageName() {
    return "ntt.Value";
  }
  private:
  inline ::PROTOBUF_NAMESPACE_ID::Arena* GetArenaNoVirtual() const {
    return nullptr;
  }
  inline void* MaybeArenaPtr() const {
    return nullptr;
  }
  public:

  ::PROTOBUF_NAMESPACE_ID::Metadata GetMetadata() const final;
  private:
  static ::PROTOBUF_NAMESPACE_ID::Metadata GetMetadataStatic() {
    ::PROTOBUF_NAMESPACE_ID::internal::AssignDescriptors(&::descriptor_table_value_2eproto);
    return ::descriptor_table_value_2eproto.file_level_metadata[kIndexInFileMessages];
  }

  public:

  // nested types ----------------------------------------------------

  // accessors -------------------------------------------------------

  enum : int {
    kByteValueFieldNumber = 1,
    kBoolValueFieldNumber = 2,
    kVerdictValueFieldNumber = 3,
    kStringValueFieldNumber = 4,
    kFloatValueFieldNumber = 5,
    kIntValueFieldNumber = 6,
    kBigValueFieldNumber = 7,
    kCompositeValueFieldNumber = 8,
  };
  // bytes byte_value = 1;
  private:
  bool _internal_has_byte_value() const;
  public:
  void clear_byte_value();
  const std::string& byte_value() const;
  void set_byte_value(const std::string& value);
  void set_byte_value(std::string&& value);
  void set_byte_value(const char* value);
  void set_byte_value(const void* value, size_t size);
  std::string* mutable_byte_value();
  std::string* release_byte_value();
  void set_allocated_byte_value(std::string* byte_value);
  private:
  const std::string& _internal_byte_value() const;
  void _internal_set_byte_value(const std::string& value);
  std::string* _internal_mutable_byte_value();
  public:

  // bool bool_value = 2;
  private:
  bool _internal_has_bool_value() const;
  public:
  void clear_bool_value();
  bool bool_value() const;
  void set_bool_value(bool value);
  private:
  bool _internal_bool_value() const;
  void _internal_set_bool_value(bool value);
  public:

  // .ntt.Verdict verdict_value = 3;
  private:
  bool _internal_has_verdict_value() const;
  public:
  void clear_verdict_value();
  ::ntt::Verdict verdict_value() const;
  void set_verdict_value(::ntt::Verdict value);
  private:
  ::ntt::Verdict _internal_verdict_value() const;
  void _internal_set_verdict_value(::ntt::Verdict value);
  public:

  // string string_value = 4;
  private:
  bool _internal_has_string_value() const;
  public:
  void clear_string_value();
  const std::string& string_value() const;
  void set_string_value(const std::string& value);
  void set_string_value(std::string&& value);
  void set_string_value(const char* value);
  void set_string_value(const char* value, size_t size);
  std::string* mutable_string_value();
  std::string* release_string_value();
  void set_allocated_string_value(std::string* string_value);
  private:
  const std::string& _internal_string_value() const;
  void _internal_set_string_value(const std::string& value);
  std::string* _internal_mutable_string_value();
  public:

  // double float_value = 5;
  private:
  bool _internal_has_float_value() const;
  public:
  void clear_float_value();
  double float_value() const;
  void set_float_value(double value);
  private:
  double _internal_float_value() const;
  void _internal_set_float_value(double value);
  public:

  // int32 int_value = 6;
  private:
  bool _internal_has_int_value() const;
  public:
  void clear_int_value();
  ::PROTOBUF_NAMESPACE_ID::int32 int_value() const;
  void set_int_value(::PROTOBUF_NAMESPACE_ID::int32 value);
  private:
  ::PROTOBUF_NAMESPACE_ID::int32 _internal_int_value() const;
  void _internal_set_int_value(::PROTOBUF_NAMESPACE_ID::int32 value);
  public:

  // string big_value = 7;
  private:
  bool _internal_has_big_value() const;
  public:
  void clear_big_value();
  const std::string& big_value() const;
  void set_big_value(const std::string& value);
  void set_big_value(std::string&& value);
  void set_big_value(const char* value);
  void set_big_value(const char* value, size_t size);
  std::string* mutable_big_value();
  std::string* release_big_value();
  void set_allocated_big_value(std::string* big_value);
  private:
  const std::string& _internal_big_value() const;
  void _internal_set_big_value(const std::string& value);
  std::string* _internal_mutable_big_value();
  public:

  // .ntt.Composite composite_value = 8;
  bool has_composite_value() const;
  private:
  bool _internal_has_composite_value() const;
  public:
  void clear_composite_value();
  const ::ntt::Composite& composite_value() const;
  ::ntt::Composite* release_composite_value();
  ::ntt::Composite* mutable_composite_value();
  void set_allocated_composite_value(::ntt::Composite* composite_value);
  private:
  const ::ntt::Composite& _internal_composite_value() const;
  ::ntt::Composite* _internal_mutable_composite_value();
  public:

  void clear_kind();
  KindCase kind_case() const;
  // @@protoc_insertion_point(class_scope:ntt.Value)
 private:
  class _Internal;
  void set_has_byte_value();
  void set_has_bool_value();
  void set_has_verdict_value();
  void set_has_string_value();
  void set_has_float_value();
  void set_has_int_value();
  void set_has_big_value();
  void set_has_composite_value();

  inline bool has_kind() const;
  inline void clear_has_kind();

  ::PROTOBUF_NAMESPACE_ID::internal::InternalMetadataWithArena _internal_metadata_;
  union KindUnion {
    KindUnion() {}
    ::PROTOBUF_NAMESPACE_ID::internal::ArenaStringPtr byte_value_;
    bool bool_value_;
    int verdict_value_;
    ::PROTOBUF_NAMESPACE_ID::internal::ArenaStringPtr string_value_;
    double float_value_;
    ::PROTOBUF_NAMESPACE_ID::int32 int_value_;
    ::PROTOBUF_NAMESPACE_ID::internal::ArenaStringPtr big_value_;
    ::ntt::Composite* composite_value_;
  } kind_;
  mutable ::PROTOBUF_NAMESPACE_ID::internal::CachedSize _cached_size_;
  ::PROTOBUF_NAMESPACE_ID::uint32 _oneof_case_[1];

  friend struct ::TableStruct_value_2eproto;
};
// -------------------------------------------------------------------

class Composite :
    public ::PROTOBUF_NAMESPACE_ID::Message /* @@protoc_insertion_point(class_definition:ntt.Composite) */ {
 public:
  Composite();
  virtual ~Composite();

  Composite(const Composite& from);
  Composite(Composite&& from) noexcept
    : Composite() {
    *this = ::std::move(from);
  }

  inline Composite& operator=(const Composite& from) {
    CopyFrom(from);
    return *this;
  }
  inline Composite& operator=(Composite&& from) noexcept {
    if (GetArenaNoVirtual() == from.GetArenaNoVirtual()) {
      if (this != &from) InternalSwap(&from);
    } else {
      CopyFrom(from);
    }
    return *this;
  }

  static const ::PROTOBUF_NAMESPACE_ID::Descriptor* descriptor() {
    return GetDescriptor();
  }
  static const ::PROTOBUF_NAMESPACE_ID::Descriptor* GetDescriptor() {
    return GetMetadataStatic().descriptor;
  }
  static const ::PROTOBUF_NAMESPACE_ID::Reflection* GetReflection() {
    return GetMetadataStatic().reflection;
  }
  static const Composite& default_instance();

  static void InitAsDefaultInstance();  // FOR INTERNAL USE ONLY
  static inline const Composite* internal_default_instance() {
    return reinterpret_cast<const Composite*>(
               &_Composite_default_instance_);
  }
  static constexpr int kIndexInFileMessages =
    1;

  friend void swap(Composite& a, Composite& b) {
    a.Swap(&b);
  }
  inline void Swap(Composite* other) {
    if (other == this) return;
    InternalSwap(other);
  }

  // implements Message ----------------------------------------------

  inline Composite* New() const final {
    return CreateMaybeMessage<Composite>(nullptr);
  }

  Composite* New(::PROTOBUF_NAMESPACE_ID::Arena* arena) const final {
    return CreateMaybeMessage<Composite>(arena);
  }
  void CopyFrom(const ::PROTOBUF_NAMESPACE_ID::Message& from) final;
  void MergeFrom(const ::PROTOBUF_NAMESPACE_ID::Message& from) final;
  void CopyFrom(const Composite& from);
  void MergeFrom(const Composite& from);
  PROTOBUF_ATTRIBUTE_REINITIALIZES void Clear() final;
  bool IsInitialized() const final;

  size_t ByteSizeLong() const final;
  const char* _InternalParse(const char* ptr, ::PROTOBUF_NAMESPACE_ID::internal::ParseContext* ctx) final;
  ::PROTOBUF_NAMESPACE_ID::uint8* _InternalSerialize(
      ::PROTOBUF_NAMESPACE_ID::uint8* target, ::PROTOBUF_NAMESPACE_ID::io::EpsCopyOutputStream* stream) const final;
  int GetCachedSize() const final { return _cached_size_.Get(); }

  private:
  inline void SharedCtor();
  inline void SharedDtor();
  void SetCachedSize(int size) const final;
  void InternalSwap(Composite* other);
  friend class ::PROTOBUF_NAMESPACE_ID::internal::AnyMetadata;
  static ::PROTOBUF_NAMESPACE_ID::StringPiece FullMessageName() {
    return "ntt.Composite";
  }
  private:
  inline ::PROTOBUF_NAMESPACE_ID::Arena* GetArenaNoVirtual() const {
    return nullptr;
  }
  inline void* MaybeArenaPtr() const {
    return nullptr;
  }
  public:

  ::PROTOBUF_NAMESPACE_ID::Metadata GetMetadata() const final;
  private:
  static ::PROTOBUF_NAMESPACE_ID::Metadata GetMetadataStatic() {
    ::PROTOBUF_NAMESPACE_ID::internal::AssignDescriptors(&::descriptor_table_value_2eproto);
    return ::descriptor_table_value_2eproto.file_level_metadata[kIndexInFileMessages];
  }

  public:

  // nested types ----------------------------------------------------

  // accessors -------------------------------------------------------

  enum : int {
    kValuesFieldNumber = 1,
  };
  // repeated .ntt.Value values = 1;
  int values_size() const;
  private:
  int _internal_values_size() const;
  public:
  void clear_values();
  ::ntt::Value* mutable_values(int index);
  ::PROTOBUF_NAMESPACE_ID::RepeatedPtrField< ::ntt::Value >*
      mutable_values();
  private:
  const ::ntt::Value& _internal_values(int index) const;
  ::ntt::Value* _internal_add_values();
  public:
  const ::ntt::Value& values(int index) const;
  ::ntt::Value* add_values();
  const ::PROTOBUF_NAMESPACE_ID::RepeatedPtrField< ::ntt::Value >&
      values() const;

  // @@protoc_insertion_point(class_scope:ntt.Composite)
 private:
  class _Internal;

  ::PROTOBUF_NAMESPACE_ID::internal::InternalMetadataWithArena _internal_metadata_;
  ::PROTOBUF_NAMESPACE_ID::RepeatedPtrField< ::ntt::Value > values_;
  mutable ::PROTOBUF_NAMESPACE_ID::internal::CachedSize _cached_size_;
  friend struct ::TableStruct_value_2eproto;
};
// ===================================================================


// ===================================================================

#ifdef __GNUC__
  #pragma GCC diagnostic push
  #pragma GCC diagnostic ignored "-Wstrict-aliasing"
#endif  // __GNUC__
// Value

// bytes byte_value = 1;
inline bool Value::_internal_has_byte_value() const {
  return kind_case() == kByteValue;
}
inline void Value::set_has_byte_value() {
  _oneof_case_[0] = kByteValue;
}
inline void Value::clear_byte_value() {
  if (_internal_has_byte_value()) {
    kind_.byte_value_.DestroyNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
    clear_has_kind();
  }
}
inline const std::string& Value::byte_value() const {
  // @@protoc_insertion_point(field_get:ntt.Value.byte_value)
  return _internal_byte_value();
}
inline void Value::set_byte_value(const std::string& value) {
  _internal_set_byte_value(value);
  // @@protoc_insertion_point(field_set:ntt.Value.byte_value)
}
inline std::string* Value::mutable_byte_value() {
  // @@protoc_insertion_point(field_mutable:ntt.Value.byte_value)
  return _internal_mutable_byte_value();
}
inline const std::string& Value::_internal_byte_value() const {
  if (_internal_has_byte_value()) {
    return kind_.byte_value_.GetNoArena();
  }
  return *&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited();
}
inline void Value::_internal_set_byte_value(const std::string& value) {
  if (!_internal_has_byte_value()) {
    clear_kind();
    set_has_byte_value();
    kind_.byte_value_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  }
  kind_.byte_value_.SetNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited(), value);
}
inline void Value::set_byte_value(std::string&& value) {
  // @@protoc_insertion_point(field_set:ntt.Value.byte_value)
  if (!_internal_has_byte_value()) {
    clear_kind();
    set_has_byte_value();
    kind_.byte_value_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  }
  kind_.byte_value_.SetNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited(), ::std::move(value));
  // @@protoc_insertion_point(field_set_rvalue:ntt.Value.byte_value)
}
inline void Value::set_byte_value(const char* value) {
  GOOGLE_DCHECK(value != nullptr);
  if (!_internal_has_byte_value()) {
    clear_kind();
    set_has_byte_value();
    kind_.byte_value_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  }
  kind_.byte_value_.SetNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited(),
      ::std::string(value));
  // @@protoc_insertion_point(field_set_char:ntt.Value.byte_value)
}
inline void Value::set_byte_value(const void* value, size_t size) {
  if (!_internal_has_byte_value()) {
    clear_kind();
    set_has_byte_value();
    kind_.byte_value_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  }
  kind_.byte_value_.SetNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited(), ::std::string(
      reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:ntt.Value.byte_value)
}
inline std::string* Value::_internal_mutable_byte_value() {
  if (!_internal_has_byte_value()) {
    clear_kind();
    set_has_byte_value();
    kind_.byte_value_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  }
  return kind_.byte_value_.MutableNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
}
inline std::string* Value::release_byte_value() {
  // @@protoc_insertion_point(field_release:ntt.Value.byte_value)
  if (_internal_has_byte_value()) {
    clear_has_kind();
    return kind_.byte_value_.ReleaseNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  } else {
    return nullptr;
  }
}
inline void Value::set_allocated_byte_value(std::string* byte_value) {
  if (has_kind()) {
    clear_kind();
  }
  if (byte_value != nullptr) {
    set_has_byte_value();
    kind_.byte_value_.UnsafeSetDefault(byte_value);
  }
  // @@protoc_insertion_point(field_set_allocated:ntt.Value.byte_value)
}

// bool bool_value = 2;
inline bool Value::_internal_has_bool_value() const {
  return kind_case() == kBoolValue;
}
inline void Value::set_has_bool_value() {
  _oneof_case_[0] = kBoolValue;
}
inline void Value::clear_bool_value() {
  if (_internal_has_bool_value()) {
    kind_.bool_value_ = false;
    clear_has_kind();
  }
}
inline bool Value::_internal_bool_value() const {
  if (_internal_has_bool_value()) {
    return kind_.bool_value_;
  }
  return false;
}
inline void Value::_internal_set_bool_value(bool value) {
  if (!_internal_has_bool_value()) {
    clear_kind();
    set_has_bool_value();
  }
  kind_.bool_value_ = value;
}
inline bool Value::bool_value() const {
  // @@protoc_insertion_point(field_get:ntt.Value.bool_value)
  return _internal_bool_value();
}
inline void Value::set_bool_value(bool value) {
  _internal_set_bool_value(value);
  // @@protoc_insertion_point(field_set:ntt.Value.bool_value)
}

// .ntt.Verdict verdict_value = 3;
inline bool Value::_internal_has_verdict_value() const {
  return kind_case() == kVerdictValue;
}
inline void Value::set_has_verdict_value() {
  _oneof_case_[0] = kVerdictValue;
}
inline void Value::clear_verdict_value() {
  if (_internal_has_verdict_value()) {
    kind_.verdict_value_ = 0;
    clear_has_kind();
  }
}
inline ::ntt::Verdict Value::_internal_verdict_value() const {
  if (_internal_has_verdict_value()) {
    return static_cast< ::ntt::Verdict >(kind_.verdict_value_);
  }
  return static_cast< ::ntt::Verdict >(0);
}
inline ::ntt::Verdict Value::verdict_value() const {
  // @@protoc_insertion_point(field_get:ntt.Value.verdict_value)
  return _internal_verdict_value();
}
inline void Value::_internal_set_verdict_value(::ntt::Verdict value) {
  if (!_internal_has_verdict_value()) {
    clear_kind();
    set_has_verdict_value();
  }
  kind_.verdict_value_ = value;
}
inline void Value::set_verdict_value(::ntt::Verdict value) {
  // @@protoc_insertion_point(field_set:ntt.Value.verdict_value)
  _internal_set_verdict_value(value);
}

// string string_value = 4;
inline bool Value::_internal_has_string_value() const {
  return kind_case() == kStringValue;
}
inline void Value::set_has_string_value() {
  _oneof_case_[0] = kStringValue;
}
inline void Value::clear_string_value() {
  if (_internal_has_string_value()) {
    kind_.string_value_.DestroyNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
    clear_has_kind();
  }
}
inline const std::string& Value::string_value() const {
  // @@protoc_insertion_point(field_get:ntt.Value.string_value)
  return _internal_string_value();
}
inline void Value::set_string_value(const std::string& value) {
  _internal_set_string_value(value);
  // @@protoc_insertion_point(field_set:ntt.Value.string_value)
}
inline std::string* Value::mutable_string_value() {
  // @@protoc_insertion_point(field_mutable:ntt.Value.string_value)
  return _internal_mutable_string_value();
}
inline const std::string& Value::_internal_string_value() const {
  if (_internal_has_string_value()) {
    return kind_.string_value_.GetNoArena();
  }
  return *&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited();
}
inline void Value::_internal_set_string_value(const std::string& value) {
  if (!_internal_has_string_value()) {
    clear_kind();
    set_has_string_value();
    kind_.string_value_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  }
  kind_.string_value_.SetNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited(), value);
}
inline void Value::set_string_value(std::string&& value) {
  // @@protoc_insertion_point(field_set:ntt.Value.string_value)
  if (!_internal_has_string_value()) {
    clear_kind();
    set_has_string_value();
    kind_.string_value_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  }
  kind_.string_value_.SetNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited(), ::std::move(value));
  // @@protoc_insertion_point(field_set_rvalue:ntt.Value.string_value)
}
inline void Value::set_string_value(const char* value) {
  GOOGLE_DCHECK(value != nullptr);
  if (!_internal_has_string_value()) {
    clear_kind();
    set_has_string_value();
    kind_.string_value_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  }
  kind_.string_value_.SetNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited(),
      ::std::string(value));
  // @@protoc_insertion_point(field_set_char:ntt.Value.string_value)
}
inline void Value::set_string_value(const char* value, size_t size) {
  if (!_internal_has_string_value()) {
    clear_kind();
    set_has_string_value();
    kind_.string_value_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  }
  kind_.string_value_.SetNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited(), ::std::string(
      reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:ntt.Value.string_value)
}
inline std::string* Value::_internal_mutable_string_value() {
  if (!_internal_has_string_value()) {
    clear_kind();
    set_has_string_value();
    kind_.string_value_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  }
  return kind_.string_value_.MutableNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
}
inline std::string* Value::release_string_value() {
  // @@protoc_insertion_point(field_release:ntt.Value.string_value)
  if (_internal_has_string_value()) {
    clear_has_kind();
    return kind_.string_value_.ReleaseNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  } else {
    return nullptr;
  }
}
inline void Value::set_allocated_string_value(std::string* string_value) {
  if (has_kind()) {
    clear_kind();
  }
  if (string_value != nullptr) {
    set_has_string_value();
    kind_.string_value_.UnsafeSetDefault(string_value);
  }
  // @@protoc_insertion_point(field_set_allocated:ntt.Value.string_value)
}

// double float_value = 5;
inline bool Value::_internal_has_float_value() const {
  return kind_case() == kFloatValue;
}
inline void Value::set_has_float_value() {
  _oneof_case_[0] = kFloatValue;
}
inline void Value::clear_float_value() {
  if (_internal_has_float_value()) {
    kind_.float_value_ = 0;
    clear_has_kind();
  }
}
inline double Value::_internal_float_value() const {
  if (_internal_has_float_value()) {
    return kind_.float_value_;
  }
  return 0;
}
inline void Value::_internal_set_float_value(double value) {
  if (!_internal_has_float_value()) {
    clear_kind();
    set_has_float_value();
  }
  kind_.float_value_ = value;
}
inline double Value::float_value() const {
  // @@protoc_insertion_point(field_get:ntt.Value.float_value)
  return _internal_float_value();
}
inline void Value::set_float_value(double value) {
  _internal_set_float_value(value);
  // @@protoc_insertion_point(field_set:ntt.Value.float_value)
}

// int32 int_value = 6;
inline bool Value::_internal_has_int_value() const {
  return kind_case() == kIntValue;
}
inline void Value::set_has_int_value() {
  _oneof_case_[0] = kIntValue;
}
inline void Value::clear_int_value() {
  if (_internal_has_int_value()) {
    kind_.int_value_ = 0;
    clear_has_kind();
  }
}
inline ::PROTOBUF_NAMESPACE_ID::int32 Value::_internal_int_value() const {
  if (_internal_has_int_value()) {
    return kind_.int_value_;
  }
  return 0;
}
inline void Value::_internal_set_int_value(::PROTOBUF_NAMESPACE_ID::int32 value) {
  if (!_internal_has_int_value()) {
    clear_kind();
    set_has_int_value();
  }
  kind_.int_value_ = value;
}
inline ::PROTOBUF_NAMESPACE_ID::int32 Value::int_value() const {
  // @@protoc_insertion_point(field_get:ntt.Value.int_value)
  return _internal_int_value();
}
inline void Value::set_int_value(::PROTOBUF_NAMESPACE_ID::int32 value) {
  _internal_set_int_value(value);
  // @@protoc_insertion_point(field_set:ntt.Value.int_value)
}

// string big_value = 7;
inline bool Value::_internal_has_big_value() const {
  return kind_case() == kBigValue;
}
inline void Value::set_has_big_value() {
  _oneof_case_[0] = kBigValue;
}
inline void Value::clear_big_value() {
  if (_internal_has_big_value()) {
    kind_.big_value_.DestroyNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
    clear_has_kind();
  }
}
inline const std::string& Value::big_value() const {
  // @@protoc_insertion_point(field_get:ntt.Value.big_value)
  return _internal_big_value();
}
inline void Value::set_big_value(const std::string& value) {
  _internal_set_big_value(value);
  // @@protoc_insertion_point(field_set:ntt.Value.big_value)
}
inline std::string* Value::mutable_big_value() {
  // @@protoc_insertion_point(field_mutable:ntt.Value.big_value)
  return _internal_mutable_big_value();
}
inline const std::string& Value::_internal_big_value() const {
  if (_internal_has_big_value()) {
    return kind_.big_value_.GetNoArena();
  }
  return *&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited();
}
inline void Value::_internal_set_big_value(const std::string& value) {
  if (!_internal_has_big_value()) {
    clear_kind();
    set_has_big_value();
    kind_.big_value_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  }
  kind_.big_value_.SetNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited(), value);
}
inline void Value::set_big_value(std::string&& value) {
  // @@protoc_insertion_point(field_set:ntt.Value.big_value)
  if (!_internal_has_big_value()) {
    clear_kind();
    set_has_big_value();
    kind_.big_value_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  }
  kind_.big_value_.SetNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited(), ::std::move(value));
  // @@protoc_insertion_point(field_set_rvalue:ntt.Value.big_value)
}
inline void Value::set_big_value(const char* value) {
  GOOGLE_DCHECK(value != nullptr);
  if (!_internal_has_big_value()) {
    clear_kind();
    set_has_big_value();
    kind_.big_value_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  }
  kind_.big_value_.SetNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited(),
      ::std::string(value));
  // @@protoc_insertion_point(field_set_char:ntt.Value.big_value)
}
inline void Value::set_big_value(const char* value, size_t size) {
  if (!_internal_has_big_value()) {
    clear_kind();
    set_has_big_value();
    kind_.big_value_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  }
  kind_.big_value_.SetNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited(), ::std::string(
      reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:ntt.Value.big_value)
}
inline std::string* Value::_internal_mutable_big_value() {
  if (!_internal_has_big_value()) {
    clear_kind();
    set_has_big_value();
    kind_.big_value_.UnsafeSetDefault(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  }
  return kind_.big_value_.MutableNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
}
inline std::string* Value::release_big_value() {
  // @@protoc_insertion_point(field_release:ntt.Value.big_value)
  if (_internal_has_big_value()) {
    clear_has_kind();
    return kind_.big_value_.ReleaseNoArena(&::PROTOBUF_NAMESPACE_ID::internal::GetEmptyStringAlreadyInited());
  } else {
    return nullptr;
  }
}
inline void Value::set_allocated_big_value(std::string* big_value) {
  if (has_kind()) {
    clear_kind();
  }
  if (big_value != nullptr) {
    set_has_big_value();
    kind_.big_value_.UnsafeSetDefault(big_value);
  }
  // @@protoc_insertion_point(field_set_allocated:ntt.Value.big_value)
}

// .ntt.Composite composite_value = 8;
inline bool Value::_internal_has_composite_value() const {
  return kind_case() == kCompositeValue;
}
inline bool Value::has_composite_value() const {
  return _internal_has_composite_value();
}
inline void Value::set_has_composite_value() {
  _oneof_case_[0] = kCompositeValue;
}
inline void Value::clear_composite_value() {
  if (_internal_has_composite_value()) {
    delete kind_.composite_value_;
    clear_has_kind();
  }
}
inline ::ntt::Composite* Value::release_composite_value() {
  // @@protoc_insertion_point(field_release:ntt.Value.composite_value)
  if (_internal_has_composite_value()) {
    clear_has_kind();
      ::ntt::Composite* temp = kind_.composite_value_;
    kind_.composite_value_ = nullptr;
    return temp;
  } else {
    return nullptr;
  }
}
inline const ::ntt::Composite& Value::_internal_composite_value() const {
  return _internal_has_composite_value()
      ? *kind_.composite_value_
      : *reinterpret_cast< ::ntt::Composite*>(&::ntt::_Composite_default_instance_);
}
inline const ::ntt::Composite& Value::composite_value() const {
  // @@protoc_insertion_point(field_get:ntt.Value.composite_value)
  return _internal_composite_value();
}
inline ::ntt::Composite* Value::_internal_mutable_composite_value() {
  if (!_internal_has_composite_value()) {
    clear_kind();
    set_has_composite_value();
    kind_.composite_value_ = CreateMaybeMessage< ::ntt::Composite >(
        GetArenaNoVirtual());
  }
  return kind_.composite_value_;
}
inline ::ntt::Composite* Value::mutable_composite_value() {
  // @@protoc_insertion_point(field_mutable:ntt.Value.composite_value)
  return _internal_mutable_composite_value();
}

inline bool Value::has_kind() const {
  return kind_case() != KIND_NOT_SET;
}
inline void Value::clear_has_kind() {
  _oneof_case_[0] = KIND_NOT_SET;
}
inline Value::KindCase Value::kind_case() const {
  return Value::KindCase(_oneof_case_[0]);
}
// -------------------------------------------------------------------

// Composite

// repeated .ntt.Value values = 1;
inline int Composite::_internal_values_size() const {
  return values_.size();
}
inline int Composite::values_size() const {
  return _internal_values_size();
}
inline void Composite::clear_values() {
  values_.Clear();
}
inline ::ntt::Value* Composite::mutable_values(int index) {
  // @@protoc_insertion_point(field_mutable:ntt.Composite.values)
  return values_.Mutable(index);
}
inline ::PROTOBUF_NAMESPACE_ID::RepeatedPtrField< ::ntt::Value >*
Composite::mutable_values() {
  // @@protoc_insertion_point(field_mutable_list:ntt.Composite.values)
  return &values_;
}
inline const ::ntt::Value& Composite::_internal_values(int index) const {
  return values_.Get(index);
}
inline const ::ntt::Value& Composite::values(int index) const {
  // @@protoc_insertion_point(field_get:ntt.Composite.values)
  return _internal_values(index);
}
inline ::ntt::Value* Composite::_internal_add_values() {
  return values_.Add();
}
inline ::ntt::Value* Composite::add_values() {
  // @@protoc_insertion_point(field_add:ntt.Composite.values)
  return _internal_add_values();
}
inline const ::PROTOBUF_NAMESPACE_ID::RepeatedPtrField< ::ntt::Value >&
Composite::values() const {
  // @@protoc_insertion_point(field_list:ntt.Composite.values)
  return values_;
}

#ifdef __GNUC__
  #pragma GCC diagnostic pop
#endif  // __GNUC__
// -------------------------------------------------------------------


// @@protoc_insertion_point(namespace_scope)

}  // namespace ntt

PROTOBUF_NAMESPACE_OPEN

template <> struct is_proto_enum< ::ntt::Verdict> : ::std::true_type {};
template <>
inline const EnumDescriptor* GetEnumDescriptor< ::ntt::Verdict>() {
  return ::ntt::Verdict_descriptor();
}

PROTOBUF_NAMESPACE_CLOSE

// @@protoc_insertion_point(global_scope)

#include <google/protobuf/port_undef.inc>
#endif  // GOOGLE_PROTOBUF_INCLUDED_GOOGLE_PROTOBUF_INCLUDED_value_2eproto
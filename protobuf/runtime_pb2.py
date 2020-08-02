# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: runtime.proto

from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


import parameter_pb2 as parameter__pb2
import value_pb2 as value__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='runtime.proto',
  package='ntt',
  syntax='proto3',
  serialized_options=b'Z\035github.com/nokia/ntt/protobuf',
  serialized_pb=b'\n\rruntime.proto\x12\x03ntt\x1a\x0fparameter.proto\x1a\x0bvalue.proto\"C\n\nRunRequest\x12\x11\n\ttest_name\x18\x01 \x01(\t\x12\"\n\nparameters\x18\x02 \x03(\x0b\x32\x0e.ntt.Parameter\"c\n\x0bRunResponse\x12\x11\n\ttest_name\x18\x01 \x01(\t\x12\"\n\nparameters\x18\x02 \x03(\x0b\x32\x0e.ntt.Parameter\x12\x1d\n\x07verdict\x18\x03 \x01(\x0e\x32\x0c.ntt.Verdict25\n\x07Runtime\x12*\n\x03Run\x12\x0f.ntt.RunRequest\x1a\x10.ntt.RunResponse\"\x00\x42\x1fZ\x1dgithub.com/nokia/ntt/protobufb\x06proto3'
  ,
  dependencies=[parameter__pb2.DESCRIPTOR,value__pb2.DESCRIPTOR,])




_RUNREQUEST = _descriptor.Descriptor(
  name='RunRequest',
  full_name='ntt.RunRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='test_name', full_name='ntt.RunRequest.test_name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='parameters', full_name='ntt.RunRequest.parameters', index=1,
      number=2, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=52,
  serialized_end=119,
)


_RUNRESPONSE = _descriptor.Descriptor(
  name='RunResponse',
  full_name='ntt.RunResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='test_name', full_name='ntt.RunResponse.test_name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='parameters', full_name='ntt.RunResponse.parameters', index=1,
      number=2, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='verdict', full_name='ntt.RunResponse.verdict', index=2,
      number=3, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=121,
  serialized_end=220,
)

_RUNREQUEST.fields_by_name['parameters'].message_type = parameter__pb2._PARAMETER
_RUNRESPONSE.fields_by_name['parameters'].message_type = parameter__pb2._PARAMETER
_RUNRESPONSE.fields_by_name['verdict'].enum_type = value__pb2._VERDICT
DESCRIPTOR.message_types_by_name['RunRequest'] = _RUNREQUEST
DESCRIPTOR.message_types_by_name['RunResponse'] = _RUNRESPONSE
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

RunRequest = _reflection.GeneratedProtocolMessageType('RunRequest', (_message.Message,), {
  'DESCRIPTOR' : _RUNREQUEST,
  '__module__' : 'runtime_pb2'
  # @@protoc_insertion_point(class_scope:ntt.RunRequest)
  })
_sym_db.RegisterMessage(RunRequest)

RunResponse = _reflection.GeneratedProtocolMessageType('RunResponse', (_message.Message,), {
  'DESCRIPTOR' : _RUNRESPONSE,
  '__module__' : 'runtime_pb2'
  # @@protoc_insertion_point(class_scope:ntt.RunResponse)
  })
_sym_db.RegisterMessage(RunResponse)


DESCRIPTOR._options = None

_RUNTIME = _descriptor.ServiceDescriptor(
  name='Runtime',
  full_name='ntt.Runtime',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  serialized_start=222,
  serialized_end=275,
  methods=[
  _descriptor.MethodDescriptor(
    name='Run',
    full_name='ntt.Runtime.Run',
    index=0,
    containing_service=None,
    input_type=_RUNREQUEST,
    output_type=_RUNRESPONSE,
    serialized_options=None,
  ),
])
_sym_db.RegisterServiceDescriptor(_RUNTIME)

DESCRIPTOR.services_by_name['Runtime'] = _RUNTIME

# @@protoc_insertion_point(module_scope)
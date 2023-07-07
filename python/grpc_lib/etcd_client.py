
import etcd_pb2_grpc as etcd
import etcd_pb2 as etcdTypes
from grpc_lib import SecureChannel
from google.protobuf.empty_pb2 import Empty

class EtcdClient(SecureChannel):
    def __init__(self, grpc_addr):
        super().__init__(grpc_addr)
        self.client = etcd.EtcdStub(self.channel)
        self.initialize_etcd()

    def initialize_etcd(self):
        empty = Empty()
        self.client.InitEtcd(empty)

    def getDatasetMetadata(self, key):
        path = etcdTypes.EtcdKey()
        path.path = key
        return self.client.GetDatasetMetadata(path)

package key

import "encoding/json"

var amiJSON = []byte(`{
  "1010.5.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-c2ca25ad",
        "hvm": "ami-cfca25a0"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-865dbfe7",
        "hvm": "ami-72ae4313"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-1f29967e",
        "hvm": "ami-c42b94a5"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-cbb47fa5",
        "hvm": "ami-83ce05ed"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-8cf17ae0",
        "hvm": "ami-038c076f"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-457d5326",
        "hvm": "ami-4b7a5428"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0905d46a",
        "hvm": "ami-d704d5b4"
      },
      {
        "name": "us-east-1",
        "pv": "ami-327e8f5f",
        "hvm": "ami-6160910c"
      },
      {
        "name": "us-west-2",
        "pv": "ami-36a95056",
        "hvm": "ami-32a85152"
      },
      {
        "name": "us-west-1",
        "pv": "ami-fc453e9c",
        "hvm": "ami-79473c19"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-706cfd03",
        "hvm": "ami-c36effb0"
      }
    ]
  },
  "1010.6.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-69c32806",
        "hvm": "ami-e6c22989"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-3f1bea5e",
        "hvm": "ami-3b18e95a"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-63ae1002",
        "hvm": "ami-b1b00ed0"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-4f76bd21",
        "hvm": "ami-8b8943e5"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-b4a237d8",
        "hvm": "ami-5aaf3a36"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-9a98b0f9",
        "hvm": "ami-25bf9746"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-dc23f1bf",
        "hvm": "ami-2c23f14f"
      },
      {
        "name": "us-east-1",
        "pv": "ami-7224ea1f",
        "hvm": "ami-9327e9fe"
      },
      {
        "name": "us-west-2",
        "pv": "ami-de1ddbbe",
        "hvm": "ami-d710d6b7"
      },
      {
        "name": "us-west-1",
        "pv": "ami-a10e49c1",
        "hvm": "ami-b20c4bd2"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-d4fb61a7",
        "hvm": "ami-e3fb6190"
      }
    ]
  },
  "1068.10.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-7b5baa14",
        "hvm": "ami-e556a78a"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-d30bc4b2",
        "hvm": "ami-df0ac5be"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-0ec8716f",
        "hvm": "ami-78c97019"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-7dba6f13",
        "hvm": "ami-f7bf6a99"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-ec5f2a83",
        "hvm": "ami-985025f7"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-989001f4",
        "hvm": "ami-4d900121"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-26dbec45",
        "hvm": "ami-2ddbec4e"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-4fbf672c",
        "hvm": "ami-e8bd658b"
      },
      {
        "name": "us-east-1",
        "pv": "ami-84e88993",
        "hvm": "ami-0aef8e1d"
      },
      {
        "name": "us-west-2",
        "pv": "ami-571cc837",
        "hvm": "ami-7d11c51d"
      },
      {
        "name": "us-west-1",
        "pv": "ami-18c48678",
        "hvm": "ami-71c48611"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-c27701b1",
        "hvm": "ami-85097ff6"
      }
    ]
  },
  "1068.6.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-e9a95c86",
        "hvm": "ami-23a85d4c"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-78ae5119",
        "hvm": "ami-2cb14e4d"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-8e7dc3ef",
        "hvm": "ami-837fc1e2"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-8bb57fe5",
        "hvm": "ami-bab77dd4"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-9cbf2bf0",
        "hvm": "ami-baba2ed6"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-77745e14",
        "hvm": "ami-a3755fc0"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-f58e5296",
        "hvm": "ami-2b8f5348"
      },
      {
        "name": "us-east-1",
        "pv": "ami-8ec74499",
        "hvm": "ami-edc744fa"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0072b160",
        "hvm": "ami-756ead15"
      },
      {
        "name": "us-west-1",
        "pv": "ami-9894d2f8",
        "hvm": "ami-a896d0c8"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-b891f7cb",
        "hvm": "ami-7292f401"
      }
    ]
  },
  "1068.8.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-e37b8e8c",
        "hvm": "ami-7b7a8f14"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-d0e21bb1",
        "hvm": "ami-fcd9209d"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-ba67d9db",
        "hvm": "ami-1d66d87c"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-9edf15f0",
        "hvm": "ami-91de14ff"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0317836f",
        "hvm": "ami-ef43d783"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-ede4ce8e",
        "hvm": "ami-e8e4ce8b"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-3203df51",
        "hvm": "ami-9b00dcf8"
      },
      {
        "name": "us-east-1",
        "pv": "ami-098e011e",
        "hvm": "ami-368c0321"
      },
      {
        "name": "us-west-2",
        "pv": "ami-ecec218c",
        "hvm": "ami-cfef22af"
      },
      {
        "name": "us-west-1",
        "pv": "ami-ae2564ce",
        "hvm": "ami-bc2465dc"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-b6b8d8c5",
        "hvm": "ami-cbb5d5b8"
      }
    ]
  },
  "1068.9.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-73e0161c",
        "hvm": "ami-3ae31555"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-bc5796dd",
        "hvm": "ami-965899f7"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-d212acb3",
        "hvm": "ami-b712acd6"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-d74b81b9",
        "hvm": "ami-a24d87cc"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-eb097c84",
        "hvm": "ami-c10b7eae"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-63e3750f",
        "hvm": "ami-61e3750d"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-3a182c59",
        "hvm": "ami-b1291dd2"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-ee21ff8d",
        "hvm": "ami-3120fe52"
      },
      {
        "name": "us-east-1",
        "pv": "ami-bf108ca8",
        "hvm": "ami-6d138f7a"
      },
      {
        "name": "us-west-2",
        "pv": "ami-f568a095",
        "hvm": "ami-dc6ba3bc"
      },
      {
        "name": "us-west-1",
        "pv": "ami-d85714b8",
        "hvm": "ami-ee57148e"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-93c9a1e0",
        "hvm": "ami-b7cba3c4"
      }
    ]
  },
  "1122.2.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-4704f728",
        "hvm": "ami-c90bf8a6"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-29fb2e48",
        "hvm": "ami-85e530e4"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-6d972e0c",
        "hvm": "ami-e6932a87"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-3916c357",
        "hvm": "ami-9014c1fe"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-8092e7ef",
        "hvm": "ami-df98edb0"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-44e47428",
        "hvm": "ami-eb27b687"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-c56150a6",
        "hvm": "ami-a86051cb"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-ab22f9c8",
        "hvm": "ami-5a26fd39"
      },
      {
        "name": "us-east-1",
        "pv": "ami-3795e020",
        "hvm": "ami-1c94e10b"
      },
      {
        "name": "us-west-2",
        "pv": "ami-daac7cba",
        "hvm": "ami-06af7f66"
      },
      {
        "name": "us-west-1",
        "pv": "ami-4a561a2a",
        "hvm": "ami-43561a23"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-58d5a82b",
        "hvm": "ami-e3d6ab90"
      }
    ]
  },
  "1122.3.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-0717ee68",
        "hvm": "ami-1809f077"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-a68325c7",
        "hvm": "ami-a38026c2"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-45d96124",
        "hvm": "ami-03da6262"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-39e13557",
        "hvm": "ami-98e733f6"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-836b1fec",
        "hvm": "ami-046e1a6b"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-45049929",
        "hvm": "ami-3a0d9056"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-8b407de8",
        "hvm": "ami-36427f55"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-9a2583f9",
        "hvm": "ami-ed23858e"
      },
      {
        "name": "us-east-1",
        "pv": "ami-bf5605a8",
        "hvm": "ami-40570457"
      },
      {
        "name": "us-west-2",
        "pv": "ami-09f55169",
        "hvm": "ami-bef450de"
      },
      {
        "name": "us-west-1",
        "pv": "ami-97074cf7",
        "hvm": "ami-17064d77"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-094f017a",
        "hvm": "ami-a94c02da"
      }
    ]
  },
  "1185.3.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-72867d1d",
        "hvm": "ami-27877c48"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-32bc1153",
        "hvm": "ami-99a30ef8"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-eda31b8c",
        "hvm": "ami-1ca61e7d"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-8427f3ea",
        "hvm": "ami-4622f628"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-cca1d5a3",
        "hvm": "ami-d4a7d3bb"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-3db82451",
        "hvm": "ami-3eb82452"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-18211d7b",
        "hvm": "ami-cd201cae"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-a8ea4acb",
        "hvm": "ami-35ec4c56"
      },
      {
        "name": "us-east-1",
        "pv": "ami-81735696",
        "hvm": "ami-4d795c5a"
      },
      {
        "name": "us-east-2",
        "pv": "ami-442d7721",
        "hvm": "ami-37217b52"
      },
      {
        "name": "us-west-2",
        "pv": "ami-1512b475",
        "hvm": "ami-6f1eb80f"
      },
      {
        "name": "us-west-1",
        "pv": "ami-1b39737b",
        "hvm": "ami-773e7417"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-b7db91c4",
        "hvm": "ami-7ddc960e"
      }
    ]
  },
  "1185.5.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-3102c45e",
        "hvm": "ami-f603c599"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-6287e505",
        "hvm": "ami-ca8be9ad"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-b2e358d3",
        "hvm": "ami-6fed560e"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-6b2ef905",
        "hvm": "ami-ce2ef9a0"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-a9b2c5c6",
        "hvm": "ami-1bb6c174"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-29ac3545",
        "hvm": "ami-38ae3754"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-46536b25",
        "hvm": "ami-4f556d2c"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-158b2776",
        "hvm": "ami-0a842869"
      },
      {
        "name": "us-east-1",
        "pv": "ami-b5e5e3a2",
        "hvm": "ami-7ee7e169"
      },
      {
        "name": "us-east-2",
        "pv": "ami-e0aef485",
        "hvm": "ami-f8aaf09d"
      },
      {
        "name": "us-west-2",
        "pv": "ami-91e64df1",
        "hvm": "ami-d0e54eb0"
      },
      {
        "name": "us-west-1",
        "pv": "ami-cedf88ae",
        "hvm": "ami-f7df8897"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-ab045ed8",
        "hvm": "ami-eb3b6198"
      }
    ]
  },
  "1235.12.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-4733f928",
        "hvm": "ami-903df7ff"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-86f1b9e1",
        "hvm": "ami-93f2baf4"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-a846fcc9",
        "hvm": "ami-e441fb85"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-0c15c562",
        "hvm": "ami-d914c4b7"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-a97bc6cd",
        "hvm": "ami-3079c454"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-4530402a",
        "hvm": "ami-a33545cc"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-c51573a9",
        "hvm": "ami-c11573ad"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-5baeae38",
        "hvm": "ami-9db0b0fe"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-27cc7d44",
        "hvm": "ami-aacc7dc9"
      },
      {
        "name": "us-east-1",
        "pv": "ami-42ad7d54",
        "hvm": "ami-1ad0000c"
      },
      {
        "name": "us-east-2",
        "pv": "ami-69a3860c",
        "hvm": "ami-42a38627"
      },
      {
        "name": "us-west-2",
        "pv": "ami-2551d145",
        "hvm": "ami-444dcd24"
      },
      {
        "name": "us-west-1",
        "pv": "ami-1a1b457a",
        "hvm": "ami-b31d43d3"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-89f6dbef",
        "hvm": "ami-abcde0cd"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-052b3e61",
        "hvm": "ami-002b3e64"
      }
    ]
  },
  "1235.4.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-9c4183f3",
        "hvm": "ami-6b478504"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-818afbe6",
        "hvm": "ami-418afb26"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-e66dd687",
        "hvm": "ami-2c61da4d"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-b30ddbdd",
        "hvm": "ami-1109df7f"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-fd18aa99",
        "hvm": "ami-0319ab67"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-39611756",
        "hvm": "ami-d76e18b8"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-5d1f8431",
        "hvm": "ami-b41b80d8"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-eb535688",
        "hvm": "ami-47505524"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-44f15927",
        "hvm": "ami-57f05834"
      },
      {
        "name": "us-east-1",
        "pv": "ami-ca312add",
        "hvm": "ami-39302b2e"
      },
      {
        "name": "us-east-2",
        "pv": "ami-6be2b80e",
        "hvm": "ami-b8e2b8dd"
      },
      {
        "name": "us-west-2",
        "pv": "ami-6b7ccc0b",
        "hvm": "ami-c177c7a1"
      },
      {
        "name": "us-west-1",
        "pv": "ami-056c3c65",
        "hvm": "ami-4f6f3f2f"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-9e6749ed",
        "hvm": "ami-70694703"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-01f2f865",
        "hvm": "ami-5ffdf73b"
      }
    ]
  },
  "1235.5.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-31ee235e",
        "hvm": "ami-f6ef2299"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-68c7b40f",
        "hvm": "ami-89c6b5ee"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-f35ce792",
        "hvm": "ami-a052e9c1"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-e04e988e",
        "hvm": "ami-a34096cd"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-4e18aa2a",
        "hvm": "ami-f401b390"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-e04c3a8f",
        "hvm": "ami-304e385f"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-b1fc66dd",
        "hvm": "ami-0ffc6663"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-c1d6d3a2",
        "hvm": "ami-41d5d022"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-ff50f89c",
        "hvm": "ami-4052fa23"
      },
      {
        "name": "us-east-1",
        "pv": "ami-e243a4f4",
        "hvm": "ami-b842a5ae"
      },
      {
        "name": "us-east-2",
        "pv": "ami-6d664308",
        "hvm": "ami-06614463"
      },
      {
        "name": "us-west-2",
        "pv": "ami-5f22913f",
        "hvm": "ami-0d21926d"
      },
      {
        "name": "us-west-1",
        "pv": "ami-a6fcacc6",
        "hvm": "ami-75feae15"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-d7bf96a4",
        "hvm": "ami-6dbc951e"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-0b8a806f",
        "hvm": "ami-89f5ffed"
      }
    ]
  },
  "1235.6.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-913bf6fe",
        "hvm": "ami-113df07e"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-fa85f79d",
        "hvm": "ami-1b8af87c"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-3941fa58",
        "hvm": "ami-3c47fc5d"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-9b7badf5",
        "hvm": "ami-7379af1d"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-e507b581",
        "hvm": "ami-9c04b6f8"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-f5a4d29a",
        "hvm": "ami-d0a6d0bf"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-e9900a85",
        "hvm": "ami-d7940ebb"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-6d8e8b0e",
        "hvm": "ami-9d888dfe"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-8911b9ea",
        "hvm": "ami-3315bd50"
      },
      {
        "name": "us-east-1",
        "pv": "ami-187c9d0e",
        "hvm": "ami-3b7f9e2d"
      },
      {
        "name": "us-east-2",
        "pv": "ami-41634624",
        "hvm": "ami-e66d4883"
      },
      {
        "name": "us-west-2",
        "pv": "ami-53942633",
        "hvm": "ami-12942672"
      },
      {
        "name": "us-west-1",
        "pv": "ami-43356623",
        "hvm": "ami-65336005"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-72dcf401",
        "hvm": "ami-1eddf56d"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-12f5ff76",
        "hvm": "ami-328a8056"
      }
    ]
  },
  "1235.8.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-9bb178f4",
        "hvm": "ami-a1b178ce"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-4696d121",
        "hvm": "ami-5c94d33b"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-83df65e2",
        "hvm": "ami-d3df65b2"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-e826f786",
        "hvm": "ami-fc29f892"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-40ca7724",
        "hvm": "ami-41ca7725"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-4e7b0a21",
        "hvm": "ami-3479085b"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-ce6e0aa2",
        "hvm": "ami-146d0978"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-37606654",
        "hvm": "ami-026f6961"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-3d0fba5e",
        "hvm": "ami-0c0cb96f"
      },
      {
        "name": "us-east-1",
        "pv": "ami-6a25db7c",
        "hvm": "ami-ec25dbfa"
      },
      {
        "name": "us-east-2",
        "pv": "ami-ff391c9a",
        "hvm": "ami-51381d34"
      },
      {
        "name": "us-west-2",
        "pv": "ami-8df348ed",
        "hvm": "ami-6df74c0d"
      },
      {
        "name": "us-west-1",
        "pv": "ami-836c31e3",
        "hvm": "ami-bc6e33dc"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-64bce402",
        "hvm": "ami-7dbee61b"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-9e584dfa",
        "hvm": "ami-cc5b4ea8"
      }
    ]
  },
  "1235.9.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-dd0fc6b2",
        "hvm": "ami-9501c8fa"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-0d51176a",
        "hvm": "ami-885f19ef"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-5ec77d3f",
        "hvm": "ami-12c67c73"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-8f5988e1",
        "hvm": "ami-d65889b8"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-c6c77aa2",
        "hvm": "ami-c8c67bac"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-a96716c6",
        "hvm": "ami-7e641511"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-285e3a44",
        "hvm": "ami-3e5d3952"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-aa2325c9",
        "hvm": "ami-d92422ba"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-15c37776",
        "hvm": "ami-14cc7877"
      },
      {
        "name": "us-east-1",
        "pv": "ami-c96d95df",
        "hvm": "ami-fd6c94eb"
      },
      {
        "name": "us-east-2",
        "pv": "ami-960326f3",
        "hvm": "ami-72032617"
      },
      {
        "name": "us-west-2",
        "pv": "ami-b74df6d7",
        "hvm": "ami-4c49f22c"
      },
      {
        "name": "us-west-1",
        "pv": "ami-ddb7eabd",
        "hvm": "ami-b6bae7d6"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-188dd67e",
        "hvm": "ami-ac8fd4ca"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-57405533",
        "hvm": "ami-054c5961"
      }
    ]
  },
  "1298.5.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-5014c13f",
        "hvm": "ami-eceb3e83"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-7cca9f1b",
        "hvm": "ami-20e7b247"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-9f209afe",
        "hvm": "ami-77249e16"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-0a25f564",
        "hvm": "ami-ea26f684"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-056cd161",
        "hvm": "ami-566fd232"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-6002720f",
        "hvm": "ami-56027239"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0785e36b",
        "hvm": "ami-f98cea95"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-2fe3e04c",
        "hvm": "ami-fde3e09e"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-19fb4b7a",
        "hvm": "ami-59e2523a"
      },
      {
        "name": "us-east-1",
        "pv": "ami-33f92725",
        "hvm": "ami-eff12ff9"
      },
      {
        "name": "us-east-2",
        "pv": "ami-2586a340",
        "hvm": "ami-e387a286"
      },
      {
        "name": "us-west-2",
        "pv": "ami-fb92109b",
        "hvm": "ami-fd92109d"
      },
      {
        "name": "us-west-1",
        "pv": "ami-3b8dd35b",
        "hvm": "ami-818bd5e1"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-75507e13",
        "hvm": "ami-4829072e"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-03c1d467",
        "hvm": "ami-2bc2d74f"
      }
    ]
  },
  "1298.6.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-2126f14e",
        "hvm": "ami-8424f3eb"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-15471a72",
        "hvm": "ami-e6461b81"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-46fb7e27",
        "hvm": "ami-83fb7ee2"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-51d5063f",
        "hvm": "ami-52d5063c"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-ea3e838e",
        "hvm": "ami-b73d80d3"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-8186f6ee",
        "hvm": "ami-3d89f952"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-33f1905f",
        "hvm": "ami-adf091c1"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-59202d3a",
        "hvm": "ami-d2232eb1"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-2fa2104c",
        "hvm": "ami-c7a210a4"
      },
      {
        "name": "us-east-1",
        "pv": "ami-78359b6e",
        "hvm": "ami-55339d43"
      },
      {
        "name": "us-east-2",
        "pv": "ami-6b57730e",
        "hvm": "ami-23527646"
      },
      {
        "name": "us-west-2",
        "pv": "ami-71ef6611",
        "hvm": "ami-70ef6610"
      },
      {
        "name": "us-west-1",
        "pv": "ami-4c28702c",
        "hvm": "ami-bf2870df"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-f0fbcc96",
        "hvm": "ami-79fccb1f"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-c7b3a6a3",
        "hvm": "ami-62b1a406"
      }
    ]
  },
  "1298.7.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-11f5257e",
        "hvm": "ami-c6f424a9"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-3903275e",
        "hvm": "ami-ad0f2bca"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-5b0c893a",
        "hvm": "ami-070f8a66"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-f263b09c",
        "hvm": "ami-2163b04f"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-df09b4bb",
        "hvm": "ami-d004b9b4"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-a06f1ccf",
        "hvm": "ami-286d1e47"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-dd1073b1",
        "hvm": "ami-c51675a9"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-12d0df71",
        "hvm": "ami-32d2dd51"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-aee956cd",
        "hvm": "ami-deeb54bd"
      },
      {
        "name": "us-east-1",
        "pv": "ami-ebb035fd",
        "hvm": "ami-6bb93c7d"
      },
      {
        "name": "us-east-2",
        "pv": "ami-a7f1d5c2",
        "hvm": "ami-40f7d325"
      },
      {
        "name": "us-west-2",
        "pv": "ami-46c35426",
        "hvm": "ami-fcc4539c"
      },
      {
        "name": "us-west-1",
        "pv": "ami-3f005a5f",
        "hvm": "ami-ef015b8f"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0ca7986a",
        "hvm": "ami-f6a49b90"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-8b1501ef",
        "hvm": "ami-16150172"
      }
    ]
  },
  "1353.6.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-50a68d37",
        "hvm": "ami-e4bb9083"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-be4391d0",
        "hvm": "ami-f441939a"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-d76416b8",
        "hvm": "ami-a66715c9"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-e3e25980",
        "hvm": "ami-2be75c48"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-eb1c1488",
        "hvm": "ami-ec1c148f"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-39b5095d",
        "hvm": "ami-a6bb07c2"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-c000d6ad",
        "hvm": "ami-a10ed8cc"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-bda07cd2",
        "hvm": "ami-bda77bd2"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-7af6f71c",
        "hvm": "ami-24fafb42"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-9eecf8fa",
        "hvm": "ami-b2e3f7d6"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-bc9ef3d0",
        "hvm": "ami-0c9ef360"
      },
      {
        "name": "us-east-1",
        "pv": "ami-dbd942cd",
        "hvm": "ami-61d84377"
      },
      {
        "name": "us-east-2",
        "pv": "ami-7b55721e",
        "hvm": "ami-dc4067b9"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-0d95116c",
        "hvm": "ami-05971364"
      },
      {
        "name": "us-west-1",
        "pv": "ami-c6b692a6",
        "hvm": "ami-83b793e3"
      },
      {
        "name": "us-west-2",
        "pv": "ami-5b61fe3b",
        "hvm": "ami-7560ff15"
      }
    ]
  },
  "1353.7.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-1885af7f",
        "hvm": "ami-8284aee5"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-2661b348",
        "hvm": "ami-b974a6d7"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-5f483a30",
        "hvm": "ami-2a403245"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-a9873cca",
        "hvm": "ami-67863d04"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-14d9d177",
        "hvm": "ami-1fc7cf7c"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-58b60a3c",
        "hvm": "ami-30b10d54"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-6e36e003",
        "hvm": "ami-6a09df07"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-2104d84e",
        "hvm": "ami-d60ad6b9"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-cbcdcdad",
        "hvm": "ami-0bcbcb6d"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-a39581c7",
        "hvm": "ami-7eeafe1a"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-eccea380",
        "hvm": "ami-a3cda0cf"
      },
      {
        "name": "us-east-1",
        "pv": "ami-a6a7c2b0",
        "hvm": "ami-ad593cbb"
      },
      {
        "name": "us-east-2",
        "pv": "ami-d62007b3",
        "hvm": "ami-102f0875"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-3ba2265a",
        "hvm": "ami-6fae2a0e"
      },
      {
        "name": "us-west-1",
        "pv": "ami-c17255a1",
        "hvm": "ami-25735445"
      },
      {
        "name": "us-west-2",
        "pv": "ami-d2aa34b2",
        "hvm": "ami-e5af3185"
      }
    ]
  },
  "1353.8.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-b48c8ad3",
        "hvm": "ami-d58284b2"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-b1c418df",
        "hvm": "ami-f7c21e99"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-024a366d",
        "hvm": "ami-80473bef"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-fe5bda9d",
        "hvm": "ami-d05bdab3"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-a26275c1",
        "hvm": "ami-1e65727d"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-5f3f833b",
        "hvm": "ami-833e82e7"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-6780510a",
        "hvm": "ami-3d815050"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-25b9634a",
        "hvm": "ami-f3b9639c"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-e07a6b86",
        "hvm": "ami-33776655"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-abdacdcf",
        "hvm": "ami-ffd9ce9b"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-f1472e9d",
        "hvm": "ami-1f432a73"
      },
      {
        "name": "us-east-1",
        "pv": "ami-75fab263",
        "hvm": "ami-a5e4acb3"
      },
      {
        "name": "us-east-2",
        "pv": "ami-3e21075b",
        "hvm": "ami-ef2e088a"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-e5bc3b84",
        "hvm": "ami-7dbc3b1c"
      },
      {
        "name": "us-west-1",
        "pv": "ami-b24764d2",
        "hvm": "ami-2a47644a"
      },
      {
        "name": "us-west-2",
        "pv": "ami-b5d4b6d5",
        "hvm": "ami-f4d4b694"
      }
    ]
  },
  "1409.2.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-55514432",
        "hvm": "ami-835643e4"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-fcd30c92",
        "hvm": "ami-69d10e07"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-3ee89751",
        "hvm": "ami-9ce59af3"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-7b65e818",
        "hvm": "ami-0365e860"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-954e5ef6",
        "hvm": "ami-494d5d2a"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-eed56a8a",
        "hvm": "ami-61d66905"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-abc716c6",
        "hvm": "ami-f0c9189d"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-747bdc1b",
        "hvm": "ami-b57ed9da"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-9da7befb",
        "hvm": "ami-a5a0b9c3"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-1d726479",
        "hvm": "ami-15716771"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-1e543f72",
        "hvm": "ami-9aabc0f6"
      },
      {
        "name": "us-east-1",
        "pv": "ami-5b6d424d",
        "hvm": "ami-4191bd57"
      },
      {
        "name": "us-east-2",
        "pv": "ami-a493b5c1",
        "hvm": "ami-c195b3a4"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-fee3659f",
        "hvm": "ami-0fee686e"
      },
      {
        "name": "us-west-1",
        "pv": "ami-49f9d429",
        "hvm": "ami-f1fad791"
      },
      {
        "name": "us-west-2",
        "pv": "ami-09b6bc70",
        "hvm": "ami-d1b7bda8"
      }
    ]
  },
  "1409.5.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-ccb5a1ab",
        "hvm": "ami-abb5a1cc"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-82ea35ec",
        "hvm": "ami-9ce936f2"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-d0314fbf",
        "hvm": "ami-f0304e9f"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-3722ae54",
        "hvm": "ami-2e23af4d"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-81cbdbe2",
        "hvm": "ami-1fcbdb7c"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-24cb7440",
        "hvm": "ami-32c97656"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-d0f928bd",
        "hvm": "ami-4ef42523"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-21cd6b4e",
        "hvm": "ami-fdcf6992"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-bd3229db",
        "hvm": "ami-523f2434"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-c5697fa1",
        "hvm": "ami-d26b7db6"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-acfc97c0",
        "hvm": "ami-8bfd96e7"
      },
      {
        "name": "us-east-1",
        "pv": "ami-e72902f1",
        "hvm": "ami-a2577cb4"
      },
      {
        "name": "us-east-2",
        "pv": "ami-5871503d",
        "hvm": "ami-20725345"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-4f81072e",
        "hvm": "ami-39830558"
      },
      {
        "name": "us-west-1",
        "pv": "ami-429fb222",
        "hvm": "ami-659cb105"
      },
      {
        "name": "us-west-2",
        "pv": "ami-bf6f7bc6",
        "hvm": "ami-1c6f7b65"
      }
    ]
  },
  "1409.6.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-287c634f",
        "hvm": "ami-b37d62d4"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-9a825cf4",
        "hvm": "ami-e7855b89"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-65c8b60a",
        "hvm": "ami-96cdb3f9"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-2f870d4c",
        "hvm": "ami-b7800ad4"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-7cddcf1f",
        "hvm": "ami-e9dcce8a"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-cf7ec1ab",
        "hvm": "ami-487ec12c"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-3a0adb57",
        "hvm": "ami-3b0adb56"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-1063c37f",
        "hvm": "ami-1f62c270"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-e2c2239b",
        "hvm": "ami-a0ff1ed9"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-64c0d600",
        "hvm": "ami-5ac2d43e"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-07097c6b",
        "hvm": "ami-f40b7e98"
      },
      {
        "name": "us-east-1",
        "pv": "ami-4f545159",
        "hvm": "ami-96494c80"
      },
      {
        "name": "us-east-2",
        "pv": "ami-4b1a3b2e",
        "hvm": "ami-c01b3aa5"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-b479ffd5",
        "hvm": "ami-44078125"
      },
      {
        "name": "us-west-1",
        "pv": "ami-53133c33",
        "hvm": "ami-e3113e83"
      },
      {
        "name": "us-west-2",
        "pv": "ami-f1150688",
        "hvm": "ami-00110279"
      }
    ]
  },
  "1409.7.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-3c9a7d5a",
        "hvm": "ami-379e7951"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-5058863e",
        "hvm": "ami-f05d839e"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-323e465d",
        "hvm": "ami-ba3048d5"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-925fcef1",
        "hvm": "ami-4c5fce2f"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-64263a07",
        "hvm": "ami-e5263a86"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-4943fc2d",
        "hvm": "ami-8442fde0"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-d24091bf",
        "hvm": "ami-ca5c8da7"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-f7399b98",
        "hvm": "ami-293f9d46"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-d9ed02a0",
        "hvm": "ami-38ef0041"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-4d514029",
        "hvm": "ami-eba8be8f"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-3794e05b",
        "hvm": "ami-e295e18e"
      },
      {
        "name": "us-east-1",
        "pv": "ami-9aecfb8c",
        "hvm": "ami-4feafd59"
      },
      {
        "name": "us-east-2",
        "pv": "ami-7297b617",
        "hvm": "ami-9995b4fc"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-55870634",
        "hvm": "ami-908203f1"
      },
      {
        "name": "us-west-1",
        "pv": "ami-1e92bc7e",
        "hvm": "ami-3093bd50"
      },
      {
        "name": "us-west-2",
        "pv": "ami-eac2dc93",
        "hvm": "ami-19c0de60"
      }
    ]
  },
  "1409.8.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-ebd5208d",
        "hvm": "ami-ccd421aa"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-eb4f9685",
        "hvm": "ami-f9419897"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-ea730985",
        "hvm": "ami-95700afa"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0cd74c6f",
        "hvm": "ami-62d14a01"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-c5b5aca6",
        "hvm": "ami-c7b5aca4"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-53cb7537",
        "hvm": "ami-b6c876d2"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-e0a9798d",
        "hvm": "ami-8ca878e1"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-6012bc0f",
        "hvm": "ami-5e15bb31"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-9841b0e1",
        "hvm": "ami-a041b0d9"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-d9e4f5bd",
        "hvm": "ami-d2e4f5b6"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-f5c4b299",
        "hvm": "ami-9dc7b1f1"
      },
      {
        "name": "us-east-1",
        "pv": "ami-aa765dd1",
        "hvm": "ami-32705b49"
      },
      {
        "name": "us-east-2",
        "pv": "ami-c3aa8aa6",
        "hvm": "ami-e1ac8c84"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-8ded6dec",
        "hvm": "ami-72ed6d13"
      },
      {
        "name": "us-west-1",
        "pv": "ami-bed8f3de",
        "hvm": "ami-cddbf0ad"
      },
      {
        "name": "us-west-2",
        "pv": "ami-a615f5de",
        "hvm": "ami-a715f5df"
      }
    ]
  },
  "1409.9.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-1322d675",
        "hvm": "ami-cd22d6ab"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-8974ade7",
        "hvm": "ami-2f77ae41"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-5c4d3733",
        "hvm": "ami-844d37eb"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-303da653",
        "hvm": "ami-273fa444"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-ec667e8f",
        "hvm": "ami-ad647cce"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-5fc27c3b",
        "hvm": "ami-37c07e53"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-25d70748",
        "hvm": "ami-88d707e5"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-fbb51c94",
        "hvm": "ami-fab21b95"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-59669620",
        "hvm": "ami-0f629276"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-db9687bf",
        "hvm": "ami-d99687bd"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-891263e5",
        "hvm": "ami-c81465a4"
      },
      {
        "name": "us-east-1",
        "pv": "ami-9189bcea",
        "hvm": "ami-268cb95d"
      },
      {
        "name": "us-east-2",
        "pv": "ami-4dbd9d28",
        "hvm": "ami-32bf9f57"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-5d97173c",
        "hvm": "ami-51961630"
      },
      {
        "name": "us-west-1",
        "pv": "ami-6386ad03",
        "hvm": "ami-8f86adef"
      },
      {
        "name": "us-west-2",
        "pv": "ami-9a24c7e2",
        "hvm": "ami-512ac929"
      }
    ]
  },
  "1465.6.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-430eff25",
        "hvm": "ami-b50cfdd3"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-c8bc64a6",
        "hvm": "ami-22be664c"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-bdd3a9d2",
        "hvm": "ami-2dd1ab42"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-26325645",
        "hvm": "ami-ed32568e"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-daa8b0b9",
        "hvm": "ami-05aab266"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-2fa9174b",
        "hvm": "ami-59a8163d"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-e6c5158b",
        "hvm": "ami-5ac91937"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-90dc74ff",
        "hvm": "ami-33df775c"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-a2a25fdb",
        "hvm": "ami-109d6069"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-eea3b28a",
        "hvm": "ami-7da3b219"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-3b9dec57",
        "hvm": "ami-239ced4f"
      },
      {
        "name": "us-east-1",
        "pv": "ami-38714c43",
        "hvm": "ami-ee774a95"
      },
      {
        "name": "us-east-2",
        "pv": "ami-d26c4fb7",
        "hvm": "ami-a57251c0"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-09b03068",
        "hvm": "ami-38ad2d59"
      },
      {
        "name": "us-west-1",
        "pv": "ami-52143e32",
        "hvm": "ami-e3ebc183"
      },
      {
        "name": "us-west-2",
        "pv": "ami-7401ec0c",
        "hvm": "ami-5106eb29"
      }
    ]
  },
  "1465.7.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-d37db9b5",
        "hvm": "ami-f771b591"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-dab66db4",
        "hvm": "ami-e9b16a87"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-3105425e",
        "hvm": "ami-820344ed"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-b0224cd3",
        "hvm": "ami-65224c06"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-37b25655",
        "hvm": "ami-01b15563"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-6f09b70b",
        "hvm": "ami-c90db3ad"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-ac24f4c1",
        "hvm": "ami-9f26f6f2"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-f4c87c9b",
        "hvm": "ami-62c97d0d"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-247abe5d",
        "hvm": "ami-417abe38"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-9db8a8f9",
        "hvm": "ami-4bb5a52f"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-873042eb",
        "hvm": "ami-2233414e"
      },
      {
        "name": "us-east-1",
        "pv": "ami-eb061690",
        "hvm": "ami-7500100e"
      },
      {
        "name": "us-east-2",
        "pv": "ami-46705223",
        "hvm": "ami-c77351a2"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-3e84075f",
        "hvm": "ami-d88605b9"
      },
      {
        "name": "us-west-1",
        "pv": "ami-c96552a9",
        "hvm": "ami-36635456"
      },
      {
        "name": "us-west-2",
        "pv": "ami-806d99f8",
        "hvm": "ami-5e6f9b26"
      }
    ]
  },
  "1465.8.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-388f465e",
        "hvm": "ami-e98c458f"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-817ca7ef",
        "hvm": "ami-2d7ca743"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-798dcb16",
        "hvm": "ami-d18dcbbe"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-8a5b2de9",
        "hvm": "ami-3f5b2d5c"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-3429cf56",
        "hvm": "ami-b02accd2"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-2e9a234a",
        "hvm": "ami-e899208c"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-19934074",
        "hvm": "ami-6292410f"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-f9d86996",
        "hvm": "ami-e1d9688e"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-1a589463",
        "hvm": "ami-40589439"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-70c7d414",
        "hvm": "ami-6cc6d508"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-61fc810d",
        "hvm": "ami-42ff822e"
      },
      {
        "name": "us-east-1",
        "pv": "ami-bbd13fc1",
        "hvm": "ami-e2d33d98"
      },
      {
        "name": "us-east-2",
        "pv": "ami-c0ba98a5",
        "hvm": "ami-5ab7953f"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-581a9939",
        "hvm": "ami-c31b98a2"
      },
      {
        "name": "us-west-1",
        "pv": "ami-277a4b47",
        "hvm": "ami-a57d4cc5"
      },
      {
        "name": "us-west-2",
        "pv": "ami-85bd41fd",
        "hvm": "ami-82bd41fa"
      }
    ]
  },
  "1520.5.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-d44e92b2",
        "hvm": "ami-80538fe6"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-f6e64398",
        "hvm": "ami-65e0450b"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-5cfaba33",
        "hvm": "ami-63f7b70c"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-6d4f370e",
        "hvm": "ami-874c34e4"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-12cc2e70",
        "hvm": "ami-f9ce2c9b"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-dd05bcb9",
        "hvm": "ami-1706bf73"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-07f4276a",
        "hvm": "ami-68f42705"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-cdff40a2",
        "hvm": "ami-cb00bca4"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-03f6257a",
        "hvm": "ami-0af22173"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-be9c8eda",
        "hvm": "ami-179c8e73"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-4c077920",
        "hvm": "ami-3707795b"
      },
      {
        "name": "us-east-1",
        "pv": "ami-4039f63a",
        "hvm": "ami-d920efa3"
      },
      {
        "name": "us-east-2",
        "pv": "ami-d495b8b1",
        "hvm": "ami-2190bd44"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-fd01839c",
        "hvm": "ami-e3018382"
      },
      {
        "name": "us-west-1",
        "pv": "ami-4d87b42d",
        "hvm": "ami-6786b507"
      },
      {
        "name": "us-west-2",
        "pv": "ami-615f9819",
        "hvm": "ami-1c509764"
      }
    ]
  },
  "1520.6.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-4701de21",
        "hvm": "ami-8f05dae9"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-dcf257b2",
        "hvm": "ami-b7ff5ad9"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-ba2665d5",
        "hvm": "ami-37256658"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-9c2d56ff",
        "hvm": "ami-602f5403"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-ba0be9d8",
        "hvm": "ami-b70be9d5"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-f0f14994",
        "hvm": "ami-6df34b09"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-63e0330e",
        "hvm": "ami-35e13258"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-ea8e3185",
        "hvm": "ami-8f8f30e0"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-8ddb09f4",
        "hvm": "ami-7edf0d07"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-d6b4a6b2",
        "hvm": "ami-efb7a58b"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-b85b25d4",
        "hvm": "ami-055b2569"
      },
      {
        "name": "us-east-1",
        "pv": "ami-b84881c2",
        "hvm": "ami-984980e2"
      },
      {
        "name": "us-east-2",
        "pv": "ami-857a56e0",
        "hvm": "ami-db7854be"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-551c9e34",
        "hvm": "ami-cb199baa"
      },
      {
        "name": "us-west-1",
        "pv": "ami-58566438",
        "hvm": "ami-6f57650f"
      },
      {
        "name": "us-west-2",
        "pv": "ami-9900c6e1",
        "hvm": "ami-d803c5a0"
      }
    ]
  },
  "1520.7.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-f2f75194",
        "hvm": "ami-a4f355c2"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-2f2f8a41",
        "hvm": "ami-872386e9"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-599ddf36",
        "hvm": "ami-1d9edc72"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-a8efa9cb",
        "hvm": "ami-81efa9e2"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-996d81fb",
        "hvm": "ami-886d81ea"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-af942ccb",
        "hvm": "ami-1c962e78"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-6c39ea01",
        "hvm": "ami-643ae909"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-03ff446c",
        "hvm": "ami-c7f942a8"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-f67ca78f",
        "hvm": "ami-e173a898"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-c1f4e9a5",
        "hvm": "ami-35f4e951"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-40dba22c",
        "hvm": "ami-56daa33a"
      },
      {
        "name": "us-east-1",
        "pv": "ami-01b56f7b",
        "hvm": "ami-b8bc66c2"
      },
      {
        "name": "us-east-2",
        "pv": "ami-20a78b45",
        "hvm": "ami-5ca48839"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-e1a92480",
        "hvm": "ami-15a82574"
      },
      {
        "name": "us-west-1",
        "pv": "ami-36a09d56",
        "hvm": "ami-8ca69bec"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0a2ae672",
        "hvm": "ami-5124e829"
      }
    ]
  },
  "1520.8.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-7d69c81b",
        "hvm": "ami-8f65c4e9"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-6b02a705",
        "hvm": "ami-5901a437"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-1fd89a70",
        "hvm": "ami-8ad89ae5"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-c4f2b3a7",
        "hvm": "ami-64f1b007"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-27896645",
        "hvm": "ami-6e89660c"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-fd853d99",
        "hvm": "ami-91853df5"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-d627f4bb",
        "hvm": "ami-d727f4ba"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-7350eb1c",
        "hvm": "ami-ea53e885"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-a5ae0bdc",
        "hvm": "ami-bbaf0ac2"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-fa908d9e",
        "hvm": "ami-c3978aa7"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-051b6369",
        "hvm": "ami-181c6474"
      },
      {
        "name": "us-east-1",
        "pv": "ami-eb9b3c91",
        "hvm": "ami-a89d3ad2"
      },
      {
        "name": "us-east-2",
        "pv": "ami-2280ac47",
        "hvm": "ami-1c81ad79"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-674dc006",
        "hvm": "ami-644dc005"
      },
      {
        "name": "us-west-1",
        "pv": "ami-cf566aaf",
        "hvm": "ami-23566a43"
      },
      {
        "name": "us-west-2",
        "pv": "ami-af4d82d7",
        "hvm": "ami-7c488704"
      }
    ]
  },
  "1520.9.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-bcb10ada",
        "hvm": "ami-26bc0740"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-3718bf59",
        "hvm": "ami-e11abd8f"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-8efcb2e1",
        "hvm": "ami-dafdb3b5"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-08acfe6b",
        "hvm": "ami-89acfeea"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-6f7e8b0d",
        "hvm": "ami-be7d88dc"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-ad44ffc9",
        "hvm": "ami-8245fee6"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-c2ea38af",
        "hvm": "ami-84e735e9"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-e8ed6287",
        "hvm": "ami-91ed62fe"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-f2bc018b",
        "hvm": "ami-f2c27f8b"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-1c6c7278",
        "hvm": "ami-6e6e700a"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-daa8ecb6",
        "hvm": "ami-bcaeead0"
      },
      {
        "name": "us-east-1",
        "pv": "ami-b27ce2c8",
        "hvm": "ami-1a7de360"
      },
      {
        "name": "us-east-2",
        "pv": "ami-f44c6591",
        "hvm": "ami-2e4f664b"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-b0ed62d1",
        "hvm": "ami-5dec633c"
      },
      {
        "name": "us-west-1",
        "pv": "ami-867b41e6",
        "hvm": "ami-ef7a408f"
      },
      {
        "name": "us-west-2",
        "pv": "ami-f760b98f",
        "hvm": "ami-1f63ba67"
      }
    ]
  },
  "1576.4.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-8a6ee9ec",
        "hvm": "ami-6a6bec0c"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-1db41273",
        "hvm": "ami-7fb41211"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-14b7fe7b",
        "hvm": "ami-02b4fd6d"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-af096dd3",
        "hvm": "ami-cb096db7"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-9754a0f5",
        "hvm": "ami-7957a31b"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-f014af94",
        "hvm": "ami-9c16adf8"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-9101d3fc",
        "hvm": "ami-e803d185"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-10c64f7f",
        "hvm": "ami-31c74e5e"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-08ad1471",
        "hvm": "ami-c8a811b1"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-89cdd3ed",
        "hvm": "ami-8ccdd3e8"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-1484c378",
        "hvm": "ami-af84c3c3"
      },
      {
        "name": "us-east-1",
        "pv": "ami-f9c1a083",
        "hvm": "ami-6dfb9a17"
      },
      {
        "name": "us-east-2",
        "pv": "ami-1ce2cb79",
        "hvm": "ami-01e2cb64"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-6aad220b",
        "hvm": "ami-6bad220a"
      },
      {
        "name": "us-west-1",
        "pv": "ami-6c83b90c",
        "hvm": "ami-7d81bb1d"
      },
      {
        "name": "us-west-2",
        "pv": "ami-3363b94b",
        "hvm": "ami-c167bdb9"
      }
    ]
  },
  "1576.5.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-2aa13d4c",
        "hvm": "ami-44a03c22"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-8bc969e5",
        "hvm": "ami-b3c969dd"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-894410e6",
        "hvm": "ami-655a0e0a"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-ef80f693",
        "hvm": "ami-d085f3ac"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-fccf3d9e",
        "hvm": "ami-21ce3c43"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-8adb5eee",
        "hvm": "ami-11e46175"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-6ad90407",
        "hvm": "ami-9bdf02f6"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-cbc754a4",
        "hvm": "ami-90c152ff"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-71da4c08",
        "hvm": "ami-32d1474b"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-ccc7dfa8",
        "hvm": "ami-f2fae296"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-7dbefd11",
        "hvm": "ami-78befd14"
      },
      {
        "name": "us-east-1",
        "pv": "ami-0984d473",
        "hvm": "ami-e582d29f"
      },
      {
        "name": "us-east-2",
        "pv": "ami-6cfad109",
        "hvm": "ami-07fbd062"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-7f77f91e",
        "hvm": "ami-9579f7f4"
      },
      {
        "name": "us-west-1",
        "pv": "ami-6e68680e",
        "hvm": "ami-e0696980"
      },
      {
        "name": "us-west-2",
        "pv": "ami-3746ec4f",
        "hvm": "ami-dc4ce6a4"
      }
    ]
  },
  "1632.2.1": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-15214b73",
        "hvm": "ami-7fdcb719"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-fe12b190",
        "hvm": "ami-2813b046"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-3bf8a854",
        "hvm": "ami-fff3a390"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-fad69286",
        "hvm": "ami-04c48078"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-6adf2708",
        "hvm": "ami-70e71f12"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-f0ec6894",
        "hvm": "ami-3ded6959"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-2065b84d",
        "hvm": "ami-1267ba7f"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-fd7be192",
        "hvm": "ami-4354ce2c"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-290f6350",
        "hvm": "ami-a22f43db"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-8eeef4ea",
        "hvm": "ami-61eaf005"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-2769264b",
        "hvm": "ami-456c2329"
      },
      {
        "name": "us-east-1",
        "pv": "ami-7be9ef01",
        "hvm": "ami-a53335df"
      },
      {
        "name": "us-east-2",
        "pv": "ami-abe2d7ce",
        "hvm": "ami-1ce1d479"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-5ed75f3f",
        "hvm": "ami-fad55d9b"
      },
      {
        "name": "us-west-1",
        "pv": "ami-026f6062",
        "hvm": "ami-35555a55"
      },
      {
        "name": "us-west-2",
        "pv": "ami-c42398bc",
        "hvm": "ami-65269d1d"
      }
    ]
  },
  "1632.3.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-024b3664",
        "hvm": "ami-884835ee"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-1554f67b",
        "hvm": "ami-1455f77a"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-a21b46cd",
        "hvm": "ami-991845f6"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-06c1837a",
        "hvm": "ami-b9c280c5"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-17b97c75",
        "hvm": "ami-04be7b66"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-ad79fdc9",
        "hvm": "ami-9e7cf8fa"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-d2d70bbf",
        "hvm": "ami-d1d70bbc"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-8f293ded"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-09224366",
        "hvm": "ami-862140e9"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-6319691a",
        "hvm": "ami-a61464df"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-a80feacf",
        "hvm": "ami-3e0eeb59"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-1f9d2b62"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-cf2f66a3",
        "hvm": "ami-022d646e"
      },
      {
        "name": "us-east-1",
        "pv": "ami-08302d72",
        "hvm": "ami-3f061b45"
      },
      {
        "name": "us-east-2",
        "pv": "ami-6ffeca0a",
        "hvm": "ami-85ffcbe0"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-880f87e9",
        "hvm": "ami-bc0d85dd"
      },
      {
        "name": "us-west-1",
        "pv": "ami-fe08019e",
        "hvm": "ami-cc0900ac"
      },
      {
        "name": "us-west-2",
        "pv": "ami-432eae3b",
        "hvm": "ami-692faf11"
      }
    ]
  },
  "1688.4.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-dd161fa1",
        "hvm": "ami-8d151cf1"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-40de712e",
        "hvm": "ami-5fdf7031"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-8be4bfe4",
        "hvm": "ami-3ee6bd51"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-2489d758",
        "hvm": "ami-4688d63a"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-93e12df1",
        "hvm": "ami-35e02c57"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-26971142",
        "hvm": "ami-878b0de3"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-d0d807bd",
        "hvm": "ami-51dc033c"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-72ffeb10"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-fc9dcf17",
        "hvm": "ami-5c9fcdb7"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-bf7020c6",
        "hvm": "ami-7e4b1b07"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-045eb863",
        "hvm": "ami-1950b67e"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0221977f"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-1006517c",
        "hvm": "ami-1a065176"
      },
      {
        "name": "us-east-1",
        "pv": "ami-a6f02fdb",
        "hvm": "ami-f5f92688"
      },
      {
        "name": "us-east-2",
        "pv": "ami-34fdcc51",
        "hvm": "ami-6cfacb09"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-50c45131",
        "hvm": "ami-91c154f0"
      },
      {
        "name": "us-west-1",
        "pv": "ami-bdfceadd",
        "hvm": "ami-8cfceaec"
      },
      {
        "name": "us-west-2",
        "pv": "ami-a50c95dd",
        "hvm": "ami-ea0a9392"
      }
    ]
  },
  "1688.5.3": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-31b1a54d",
        "hvm": "ami-a2b6a2de"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-6249e60c",
        "hvm": "ami-cd4de2a3"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-d82104b7",
        "hvm": "ami-0227026d"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-88471df4",
        "hvm": "ami-41461c3d"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0e8a446c",
        "hvm": "ami-f58e4097"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-3f58de5b",
        "hvm": "ami-7966e01d"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-62eb340f",
        "hvm": "ami-39ee3154"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-e7958185"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-084e11e3",
        "hvm": "ami-604e118b"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-405b0439",
        "hvm": "ami-34237c4d"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-dc35d4bb",
        "hvm": "ami-b530d1d2"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-a918aed4"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-6b257307",
        "hvm": "ami-8d2472e1"
      },
      {
        "name": "us-east-1",
        "pv": "ami-12298a6f",
        "hvm": "ami-9e2685e3"
      },
      {
        "name": "us-east-2",
        "pv": "ami-256f5f40",
        "hvm": "ami-5d6e5e38"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-8ea83def",
        "hvm": "ami-e0aa3f81"
      },
      {
        "name": "us-west-1",
        "pv": "ami-9cabbafc",
        "hvm": "ami-07a6b767"
      },
      {
        "name": "us-west-2",
        "pv": "ami-f81c7880",
        "hvm": "ami-b41377cc"
      }
    ]
  },
  "1745.3.1": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-3c45b843",
        "hvm": "ami-d948b5a6"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-d80ba3b6",
        "hvm": "ami-9709a1f9"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-61cfe30e",
        "hvm": "ami-bfcfe3d0"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-08714c74",
        "hvm": "ami-b5714cc9"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-7e5e8e1c",
        "hvm": "ami-a14090c3"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-912babf5",
        "hvm": "ami-d62dadb2"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-da13cdb7",
        "hvm": "ami-9712ccfa"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-d1bda9b3"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-5bba92b0",
        "hvm": "ami-6abd9581"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-24f3cc5d",
        "hvm": "ami-62e3dc1b"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-6db05c0a",
        "hvm": "ami-2eba5649"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-6251e01f"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-f9e8b795",
        "hvm": "ami-15e6b979"
      },
      {
        "name": "us-east-1",
        "pv": "ami-27492c58",
        "hvm": "ami-844f2afb"
      },
      {
        "name": "us-east-2",
        "pv": "ami-ebf1cd8e",
        "hvm": "ami-edf2ce88"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-4b76e02a",
        "hvm": "ami-2377e142"
      },
      {
        "name": "us-west-1",
        "pv": "ami-e2b6ae82",
        "hvm": "ami-79b2aa19"
      },
      {
        "name": "us-west-2",
        "pv": "ami-018ef079",
        "hvm": "ami-4e8ff136"
      }
    ]
  },
  "1745.4.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-31d9264e",
        "hvm": "ami-21d9265e"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-68ef4406",
        "hvm": "ami-efe94281"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-5d98b432",
        "hvm": "ami-0799b568"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-15ac9169",
        "hvm": "ami-73b28f0f"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-4cbe6d2e",
        "hvm": "ami-8fbf6ced"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-0a3dbd6e",
        "hvm": "ami-fb39b99f"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-8c05dbe1",
        "hvm": "ami-8f05dbe2"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-00b8ac62"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-a61b304d",
        "hvm": "ami-32042fd9"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-24675e5d",
        "hvm": "ami-82645dfb"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-beec00d9",
        "hvm": "ami-be967ad9"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-8d6cddf0"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-892d73e5",
        "hvm": "ami-c32d73af"
      },
      {
        "name": "us-east-1",
        "pv": "ami-8cd1b6f3",
        "hvm": "ami-93d3b4ec"
      },
      {
        "name": "us-east-2",
        "pv": "ami-e1cdf184",
        "hvm": "ami-e5cdf180"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-e460f685",
        "hvm": "ami-2960f648"
      },
      {
        "name": "us-west-1",
        "pv": "ami-a86378c8",
        "hvm": "ami-5e63783e"
      },
      {
        "name": "us-west-2",
        "pv": "ami-244f365c",
        "hvm": "ami-574f362f"
      }
    ]
  },
  "1745.5.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-69e51f16",
        "hvm": "ami-55e41e2a"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-aa903bc4",
        "hvm": "ami-c09338ae"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-ae5e70c1",
        "hvm": "ami-84406eeb"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-19c1ff65",
        "hvm": "ami-86bf81fa"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-b7e537d5",
        "hvm": "ami-f6e53794"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-06fa7962",
        "hvm": "ami-6df57609"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-fc5c8291",
        "hvm": "ami-555a8438"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-06a0b464"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-8680b56d",
        "hvm": "ami-4a83b6a1"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-2572775c",
        "hvm": "ami-c70005be"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-db7a96bc",
        "hvm": "ami-177a9670"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-d240f1af"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-9e267ff2",
        "hvm": "ami-a82079c4"
      },
      {
        "name": "us-east-1",
        "pv": "ami-6b3e5514",
        "hvm": "ami-a32d46dc"
      },
      {
        "name": "us-east-2",
        "pv": "ami-e4487781",
        "hvm": "ami-36497653"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-54cf5e35",
        "hvm": "ami-8ccc5ded"
      },
      {
        "name": "us-west-1",
        "pv": "ami-161a0076",
        "hvm": "ami-6e647e0e"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0e90ea76",
        "hvm": "ami-4296ec3a"
      }
    ]
  },
  "1745.6.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-872be6f8",
        "hvm": "ami-ac6babd3"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-4eb91320",
        "hvm": "ami-a8f15bc6"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-2fedc540",
        "hvm": "ami-2cb09943"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-f8101484",
        "hvm": "ami-79caf005"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-6a75a908",
        "hvm": "ami-43b56921"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-fb6dee9f",
        "hvm": "ami-714ccf15"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-fd914890",
        "hvm": "ami-a69e47cb"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-cba9bda9"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-d40b3b3f",
        "hvm": "ami-5f4d7cb4"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-75e7e50c",
        "hvm": "ami-74e6e60d"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-f258b795",
        "hvm": "ami-820be4e5"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-4c229331"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-7f0f5413",
        "hvm": "ami-a092cacc"
      },
      {
        "name": "us-east-1",
        "pv": "ami-a12763de",
        "hvm": "ami-fd3b4582"
      },
      {
        "name": "us-east-2",
        "pv": "ami-873806e2",
        "hvm": "ami-7080bf15"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-a035a4c1",
        "hvm": "ami-4e34a52f"
      },
      {
        "name": "us-west-1",
        "pv": "ami-ad8266ce",
        "hvm": "ami-2501e446"
      },
      {
        "name": "us-west-2",
        "pv": "ami-121c5d6a",
        "hvm": "ami-401f5e38"
      }
    ]
  },
  "1745.7.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-d633fca9",
        "hvm": "ami-ab20efd4"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-5c6cc632",
        "hvm": "ami-e46dc78a"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-2378534c",
        "hvm": "ami-807a51ef"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0ee1e672",
        "hvm": "ami-3cded940"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-d51bc4b7",
        "hvm": "ami-961ac5f4"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-1b088b7f",
        "hvm": "ami-f90a899d"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-c5875ea8",
        "hvm": "ami-2a875e47"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-c9abbfab"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-b3dfec58",
        "hvm": "ami-d0dcef3b"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-8ed4d8f7",
        "hvm": "ami-1ed8d467"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-c8907eaf",
        "hvm": "ami-40907e27"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-6912a314"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-3b9dc657",
        "hvm": "ami-619ec50d"
      },
      {
        "name": "us-east-1",
        "pv": "ami-50f4b42f",
        "hvm": "ami-f6ecac89"
      },
      {
        "name": "us-east-2",
        "pv": "ami-e5f0ce80",
        "hvm": "ami-5bf4ca3e"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-58cc5c39",
        "hvm": "ami-5fcc5c3e"
      },
      {
        "name": "us-west-1",
        "pv": "ami-2801e54b",
        "hvm": "ami-d90ce8ba"
      },
      {
        "name": "us-west-2",
        "pv": "ami-f0480a88",
        "hvm": "ami-662f6d1e"
      }
    ]
  },
  "1800.4.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-0197e3ec",
        "hvm": "ami-4da9dda0"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-c778cfa9",
        "hvm": "ami-9766d1f9"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-882b19e7",
        "hvm": "ami-16261479"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-40296faa",
        "hvm": "ami-1e3573f4"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-52fe5830",
        "hvm": "ami-37f85e55"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-18a32e7c",
        "hvm": "ami-39bd305d"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-86a179eb",
        "hvm": "ami-85a179e8"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-047d6a66"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-d99d9e32",
        "hvm": "ami-b49f9c5f"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-126b71f8",
        "hvm": "ami-bcacb756"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-97f71df0",
        "hvm": "ami-b8cb21df"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-2a72c257"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-230a2b4f",
        "hvm": "ami-4006272c"
      },
      {
        "name": "us-east-1",
        "pv": "ami-ed848992",
        "hvm": "ami-928885ed"
      },
      {
        "name": "us-east-2",
        "pv": "ami-410e3424",
        "hvm": "ami-b1013bd4"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-4137a520",
        "hvm": "ami-4037a521"
      },
      {
        "name": "us-west-1",
        "pv": "ami-ae1ef3cd",
        "hvm": "ami-4119f422"
      },
      {
        "name": "us-west-2",
        "pv": "ami-cf81d8b7",
        "hvm": "ami-be8ed7c6"
      }
    ]
  },
  "1800.5.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-b8661755",
        "hvm": "ami-e8f88905"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-963d8af8",
        "hvm": "ami-943d8afa"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-cc4a77a3",
        "hvm": "ami-4c695423"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-140a4afe",
        "hvm": "ami-ff084815"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-bc00a6de",
        "hvm": "ami-8e02a4ec"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-1c45c878",
        "hvm": "ami-0f44c96b"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-8dd40ce0",
        "hvm": "ami-8cd40ce1"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-1b7f6879"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-641b168f",
        "hvm": "ami-2a1518c1"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-8dd03460",
        "hvm": "ami-1d3cd8f0"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-9e1ef4f9",
        "hvm": "ami-dd0fe5ba"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-3962d244"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-bf82a3d3",
        "hvm": "ami-8d81a0e1"
      },
      {
        "name": "us-east-1",
        "pv": "ami-236e645c",
        "hvm": "ami-ab6963d4"
      },
      {
        "name": "us-east-2",
        "pv": "ami-48b9832d",
        "hvm": "ami-e0be8485"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-a0ea77c1",
        "hvm": "ami-4ae9742b"
      },
      {
        "name": "us-west-1",
        "pv": "ami-94eb07f7",
        "hvm": "ami-3cea065f"
      },
      {
        "name": "us-west-2",
        "pv": "ami-4d603b35",
        "hvm": "ami-256c375d"
      }
    ]
  },
  "1800.6.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-e8dea205",
        "hvm": "ami-e2dca00f"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-08871d16d557a6aeb",
        "hvm": "ami-04030c62eff91ed37"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-00f7eba860b4fe72e",
        "hvm": "ami-0a40e2443e565f3f6"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-b5f8ba5f",
        "hvm": "ami-6ef9bb84"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-facf6f98",
        "hvm": "ami-e8d0708a"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-ae64e9ca",
        "hvm": "ami-4560ed21"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-65ec3408",
        "hvm": "ami-afea32c2"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-7b607719"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-b0959a5b",
        "hvm": "ami-879b946c"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-071e720f28e4a7457",
        "hvm": "ami-012afb51d9c2d918f"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-758c7912",
        "hvm": "ami-289b6e4f"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0b9727badec366ad9"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0aabe44672435f9ed",
        "hvm": "ami-0b098d9d561172f16"
      },
      {
        "name": "us-east-1",
        "pv": "ami-a5fee2da",
        "hvm": "ami-b8ccd0c7"
      },
      {
        "name": "us-east-2",
        "pv": "ami-05465556e2add9edb",
        "hvm": "ami-04d978e741ee88c5d"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-c3f26ea2",
        "hvm": "ami-9dc65afc"
      },
      {
        "name": "us-west-1",
        "pv": "ami-7e917e1d",
        "hvm": "ami-55937c36"
      },
      {
        "name": "us-west-2",
        "pv": "ami-b7e2c6cf",
        "hvm": "ami-2de0c455"
      }
    ]
  },
  "1800.7.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-0f3b3f96d6f1f5813",
        "hvm": "ami-0c381f2f14ecc78f7"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-042ecc1aea82a6d89",
        "hvm": "ami-0ee8ead7a56410bf9"
      },
      {
        "name": "ap-south-1",
        "pv": "ami-08e33c0290834a0f0",
        "hvm": "ami-07a2cf92cba794c86"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-082a2acf4ee010d3d",
        "hvm": "ami-019daec4f68b5010b"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-05e4259210f2e3e46",
        "hvm": "ami-054c3b6bd4def7efd"
      },
      {
        "name": "ca-central-1",
        "pv": "ami-3c0a8758",
        "hvm": "ami-3423ae50"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-065d72bfa85c329df",
        "hvm": "ami-0d5ec5d735beb907e"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-02a5768104b4e8d4c"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-0b3f92d614771ab75",
        "hvm": "ami-03ee0a0310474a00e"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-033bd8dfbf291b96e",
        "hvm": "ami-02e92935e00c60cf0"
      },
      {
        "name": "eu-west-2",
        "pv": "ami-028148f2e4c17b394",
        "hvm": "ami-00985bd8806d05c41"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-004dbd1511bc95349"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0d6d56fe0af98c896",
        "hvm": "ami-04a7ebb302d65e89c"
      },
      {
        "name": "us-east-1",
        "pv": "ami-0dc1631ce2e20d87a",
        "hvm": "ami-00cc4337762ba4a52"
      },
      {
        "name": "us-east-2",
        "pv": "ami-0e4d690301722bb12",
        "hvm": "ami-0bb4f3b6a361ca725"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-2cc8574d",
        "hvm": "ami-1fc6597e"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0144724c13a9aeb5a",
        "hvm": "ami-03a95800c211ab99d"
      },
      {
        "name": "us-west-2",
        "pv": "ami-028e8238ae839c27d",
        "hvm": "ami-09e088627f26fd7ec"
      }
    ]
  },
  "1855.4.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-03620b047f1abddfc",
        "hvm": "ami-086eb64b7f4485a72"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-085e4381942bede7d"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-0e920ea4c7e29a7ed"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-06288d6d92925d99f",
        "hvm": "ami-0b47d43598dba794f"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0bd818523947fb872",
        "hvm": "ami-0609bb67692e98973"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-0a984ec3ead59581c"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-0253e9c990f50fbae",
        "hvm": "ami-00307a08b617fb95f"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0c0a607177f68f8c4"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-04712da793b52e462",
        "hvm": "ami-0b088568a857b7c27"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0d9ae8e150fb07db3",
        "hvm": "ami-099b2d1bdd27b4649"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-02de9d47add3bab7c"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0b8c0daca01d23eaa"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-042f154c1158c2734",
        "hvm": "ami-09ffd65c1a16012de"
      },
      {
        "name": "us-east-1",
        "pv": "ami-09db813f5ab5909c8",
        "hvm": "ami-08eda98e6fe1f83d6"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-093e794c03f1534e4"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-51dd4430",
        "hvm": "ami-98db42f9"
      },
      {
        "name": "us-west-1",
        "pv": "ami-02fa2648a4e363a96",
        "hvm": "ami-0a86d340ea7fde077"
      },
      {
        "name": "us-west-2",
        "pv": "ami-05da9819ce9ed9159",
        "hvm": "ami-02dea79d6a7f53d15"
      }
    ]
  },
  "1855.5.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-07354f1600594202f",
        "hvm": "ami-0b0fc6983c9be8f9e"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-09143ca0b3755b428"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-03eba32062e159d3c"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0c173ab56d52fc5b7",
        "hvm": "ami-0d0079786d2ee66ae"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-08bfa41e68aea0b4b",
        "hvm": "ami-029d8ef1a02553d2d"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-09985fec721ff6f89"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-0c7ca3e5b40a30dbb",
        "hvm": "ami-0211d60ca1aaa3c7d"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0deaa8ada18aec612"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-0eea1f18f973a12a6",
        "hvm": "ami-0e6601a88a9753474"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0413202555080b4f3",
        "hvm": "ami-06c40d1010f762df9"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-0f55bf46b69b768bd"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0a2c50627b8b434df"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-086b38bae87500b05",
        "hvm": "ami-0c188431919e8b2f1"
      },
      {
        "name": "us-east-1",
        "pv": "ami-035b4192e9aecfc61",
        "hvm": "ami-0f74e41ea6c13f74b"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-0ba531e8a11f8965d"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-105fc571",
        "hvm": "ami-9d58c2fc"
      },
      {
        "name": "us-west-1",
        "pv": "ami-05aa0cefb0360042d",
        "hvm": "ami-08b3af8ec59b84ef9"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0e04b96a3697ed93e",
        "hvm": "ami-05507591f0fcb2b75"
      }
    ]
  },
  "1911.3.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-01d2013a1959a897b",
        "hvm": "ami-0113626a2260333a1"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-06e6f1d3067f1de22"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-027a306ae15c48b3a"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0ac33e2f47e46d906",
        "hvm": "ami-006b1ac0974ddc205"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0225fd6c7f95404cf",
        "hvm": "ami-06ed1e9afe9c51d42"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-0181098d6495cf88d"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-0dbb58ff383e5a3d9",
        "hvm": "ami-01bd6c15d7a3e9aaa"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-04c22e04ed0182d0f"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-05bd9782de0dd6a3e",
        "hvm": "ami-0da12413bc1e32107"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0144f0b0f8721ff55",
        "hvm": "ami-0a9456e274662f3bd"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-00a945bef5cfc7054"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-074dbaf11687e8544"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0b1aece2e6af8c795",
        "hvm": "ami-04dcc70d03e772417"
      },
      {
        "name": "us-east-1",
        "pv": "ami-056910ee8d491d1f5",
        "hvm": "ami-03ed1c12a1dd84320"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-01be41f5c5bfcd6cc"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-3c9efb5d",
        "hvm": "ami-a5e580c4"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0fc0f73805fe30555",
        "hvm": "ami-0afb60118573e1488"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0f994f7efbe2bd36c",
        "hvm": "ami-0aee4947be233f88c"
      }
    ]
  },
  "1911.4.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-0ee6bb58fc140e531",
        "hvm": "ami-0bc12b4d1219f58ac"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-06da21e4dbddf08fb"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-06b7a29e2a33cb452"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0d0562e879d70e9bd",
        "hvm": "ami-0e6d89f4818a7f42c"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0eacb626044e7fd3a",
        "hvm": "ami-00a7f2d4e72882d65"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-0dd4a413d76f0b772"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-0200e438e48c94b1a",
        "hvm": "ami-0742010bbaf2d247f"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0249866ccbeab07ae"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-03538f2134836a375",
        "hvm": "ami-05e7f7fc79cd6ba7e"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-08a458503dc3340e3",
        "hvm": "ami-0772233ad155871ff"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-0016c65679adc75f5"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-015b1578841b2e1cb"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-05fdaedb150c834f3",
        "hvm": "ami-02aaf77da1aafd541"
      },
      {
        "name": "us-east-1",
        "pv": "ami-042df7b643addf6cc",
        "hvm": "ami-0f51520e8e4a1fbe7"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-08e0a720053fb44b9"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-52e28533",
        "hvm": "ami-95e582f4"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0fb454033feb80d24",
        "hvm": "ami-0daa79f4415db181c"
      },
      {
        "name": "us-west-2",
        "pv": "ami-07cbcc1b4bfb5c040",
        "hvm": "ami-0c24d5499b254c53e"
      }
    ]
  },
  "1911.5.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-0c8c9aece6c483f87",
        "hvm": "ami-09332982544efbe54"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-00bf9ba503dbbd4d5"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-0bfee97b418dee5b4"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0ca0f514ad8926278",
        "hvm": "ami-075ea608b566f4af1"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0e0c587354ed10c4c",
        "hvm": "ami-0c38e2a3786719087"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-03ccdc83e075bb995"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-026cddb82a8c43dc5",
        "hvm": "ami-0a957d1aa321cd9b0"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-01c805b12c4107d92"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-09218b2b495958238",
        "hvm": "ami-069068dd418875c0e"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-07bfd18e8e61ec667",
        "hvm": "ami-030d1bab626c90e46"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-0dd67ed6c3c8d308d"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0c9fec65b739b412c"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-05b705fa54bda529c",
        "hvm": "ami-0b693cca65d968a93"
      },
      {
        "name": "us-east-1",
        "pv": "ami-0ca5d662f9988f279",
        "hvm": "ami-0b1db01d775d666c2"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-0fd1c9a63f8fdb8a5"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-74cfae15",
        "hvm": "ami-75cfae14"
      },
      {
        "name": "us-west-1",
        "pv": "ami-03f837ea3731371fb",
        "hvm": "ami-01f0ebd86b12fe033"
      },
      {
        "name": "us-west-2",
        "pv": "ami-08dcf11eae23b7734",
        "hvm": "ami-0b5fe761216cc15dd"
      }
    ]
  },
  "1967.3.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-07bac8d675b35e587",
        "hvm": "ami-0b53d79c2d092ce77"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-0a07af788e926e417"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-06103b4b4747ad077"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-02c7f68bd43bb3cef",
        "hvm": "ami-0b0cfecc8175d06c6"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0dbc10b5920f7e072",
        "hvm": "ami-03ec12353f77620c4"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-07ec21c807d0c5176"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-02959c79557941cc1",
        "hvm": "ami-054d6909d91900dcf"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0bf16b12c53c5e9a3"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-02d1a2f9c9d3bfae5",
        "hvm": "ami-0cfe08bc0484060f5"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-025424961938fb048",
        "hvm": "ami-081f4de2b0f6032c6"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-0a937eebe9caafbb1"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0bf3ed517b76f5c0d"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0d1e38ba1866b81f4",
        "hvm": "ami-09359810f00df88f2"
      },
      {
        "name": "us-east-1",
        "pv": "ami-04cd3bf19e1cb48ed",
        "hvm": "ami-0547d9705af5e8fb2"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-08cb94999d76c8e42"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-ecc1a28d",
        "hvm": "ami-00c3a061"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0944222061c50230a",
        "hvm": "ami-0b349483829492b39"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0c32a360898f214ca",
        "hvm": "ami-0c1cc1260c7828fcb"
      }
    ]
  },
  "1967.4.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-0a73a67dbabbaf1b7",
        "hvm": "ami-0df35781c0a495bf7"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-0c118370a26168e1c"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-032ff048b7fa8575d"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0d4272ad936e91548",
        "hvm": "ami-0658741abbe1933b9"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0763c136aee710c1f",
        "hvm": "ami-016b129ed2f04127b"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-08483de146aa2cb71"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-01b66bc9fe9514562",
        "hvm": "ami-016e6c42460d72a0b"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0f0d52d85be9e0736"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-0812c73e5d6eefa9d",
        "hvm": "ami-02e8612b42a40844a"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-08444fcaab4df84c5",
        "hvm": "ami-02d7f55d7813eca77"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-0ed3713ba6554a4c6"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0cf533204e4944c3a"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0f6d623a1e019161c",
        "hvm": "ami-02c8c7d1ba3805352"
      },
      {
        "name": "us-east-1",
        "pv": "ami-08698c0766b76db4c",
        "hvm": "ami-03b89dc7d999dd591"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-08cd50e7b78b11333"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-d1117db0",
        "hvm": "ami-30177b51"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0d1735ab8d1c665bf",
        "hvm": "ami-016a53ba357781844"
      },
      {
        "name": "us-west-2",
        "pv": "ami-02bd2a9a6d8de37e6",
        "hvm": "ami-048789452aff936df"
      }
    ]
  },
  "1967.5.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-0ddefcc8e07cece6c",
        "hvm": "ami-07821bd0ea86d4511"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-01b5d118690d7c4db"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-09642e32f99945765"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-004706fb867071fbf",
        "hvm": "ami-07739b17529e8c1d0"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0bcf59418f0bc4a7c",
        "hvm": "ami-02d7d488d701a460e"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-0edacf783a84b0986"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-04ad7fb35410d72f4",
        "hvm": "ami-0d405143e313ec9cb"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0eb5198a7b6239a05"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-0ed9ac432ad3c0e7f",
        "hvm": "ami-0f46c2ed46d8157aa"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0543c8c48e1e7bfa2",
        "hvm": "ami-0628e483315b5d17e"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-0ded15c0d8a34dad2"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0dea870ebbbd767e4"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0f408213cb618db0e",
        "hvm": "ami-0d28afc45b6f88ba4"
      },
      {
        "name": "us-east-1",
        "pv": "ami-04022aa9668dfe53a",
        "hvm": "ami-0c6731558e5ca73f6"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-05df30c25dffa0eaf"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-fb7c129a",
        "hvm": "ami-07630d66"
      },
      {
        "name": "us-west-1",
        "pv": "ami-076835db72e33320d",
        "hvm": "ami-0aaec419396da3b37"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0912e959bc782a925",
        "hvm": "ami-0ac262621e0cc606d"
      }
    ]
  },
  "1967.6.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-0acc57eb804d615ab",
        "hvm": "ami-0674bd656e5bbd940"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-0dedec04918e56116"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-0bdbad103cc31c037"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-03e4eacd32a7064b2",
        "hvm": "ami-0e11019a200802b43"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-065bb02b9292abbca",
        "hvm": "ami-094bde83db4642610"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-0c119337e0f202885"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-0ad6f04897cfca179",
        "hvm": "ami-001e6f29a899df749"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-00a0d2ef649391775"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-07d47a29bae27ecb4",
        "hvm": "ami-00946a0f23931daac"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0b17d16791d3faa15",
        "hvm": "ami-0cdf1816f4d8d634e"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-0bf0bc4adb43e8fc7"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0a931bb3434fe57f0"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-09db09e8b1c42b88c",
        "hvm": "ami-0c33cc9b83b72fae6"
      },
      {
        "name": "us-east-1",
        "pv": "ami-024f4d044b1f7e4fc",
        "hvm": "ami-0089347d530e1f3e6"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-0b15c21563ba827f2"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-ae5937cf",
        "hvm": "ami-c45638a5"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0cfd5901e6956c6b0",
        "hvm": "ami-09e198d9d9ef8052b"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0e2ce0f3a14fdc3c4",
        "hvm": "ami-0b0f4f5f0c8c1a797"
      }
    ]
  },
  "2023.4.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-0fed97c6bc2803540",
        "hvm": "ami-003b3a37a48d799cf"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-0c2d3bd39b13c3b2d"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-0bd5eb3e67407e0df"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-01229fce434e832ef",
        "hvm": "ami-07aafbd1f2a182cd4"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-04c191937938729a5",
        "hvm": "ami-0cb589c5f6134f078"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-0952a9471ff71919e"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-0e75161ec484e7f70",
        "hvm": "ami-0caaf17a3032c1b56"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0a863f3b0a0720e6a"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-035c8c66c1fb3ef25",
        "hvm": "ami-015e6cb33a709348e"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-0fc64edffdfc5e7e5"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0b9e1ab008f197fcb",
        "hvm": "ami-04d747d892ccd652a"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-056a316ba69c9d9e8"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-026d41122f47f745e"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-01d6607f3dcc43b4b",
        "hvm": "ami-0e9521088a80c2a02"
      },
      {
        "name": "us-east-1",
        "pv": "ami-0d746e5b97cd483a2",
        "hvm": "ami-09d5d3bcd3e0e5c30"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-02accfa372062664b"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-007dd78814dd81561"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-867810e7",
        "hvm": "ami-07600866"
      },
      {
        "name": "us-west-1",
        "pv": "ami-002e24cafbfef268d",
        "hvm": "ami-0481a60675f6ea007"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0a41f7dc1d6b31214",
        "hvm": "ami-025acbb0fb1db6a27"
      }
    ]
  },
  "2023.5.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-0e73b03b4618a01b8",
        "hvm": "ami-0d3a9785820124591"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-03230b2fa6af112bf"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-0b85fd1356963d2ee"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-02ce7fedf82524d74",
        "hvm": "ami-0f8a9aa9857d8af7e"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-09abb449143dcee62",
        "hvm": "ami-0e87752a1d331823a"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-0c0100bac23bb1d39"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-0312487c765c2ae3c",
        "hvm": "ami-01e99c7e0a343d325"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0773341917796083a"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-03dca803621ba56df",
        "hvm": "ami-012abdf0d2781f0a5"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-09fbda19ac2fc6c3f"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0c5c18bab9e57abec",
        "hvm": "ami-01f5fbceb7a9fa4d0"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-069966bea0809e21d"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0194c504244182155"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0b9768ca5a526da2d",
        "hvm": "ami-0cd830cc037613a7d"
      },
      {
        "name": "us-east-1",
        "pv": "ami-06200cebbb5eb506f",
        "hvm": "ami-08e58b93705fb503f"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-03172282aaa2899be"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-0ff9e298ea0bacf53"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-e9f49f88",
        "hvm": "ami-e7f59e86"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0fe069c504547db88",
        "hvm": "ami-08d3e245ebf4d560f"
      },
      {
        "name": "us-west-2",
        "pv": "ami-08e9621af018d03ad",
        "hvm": "ami-0a4f49b2488e15346"
      }
    ]
  },
  "2079.3.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-0ec0449c92b8678c5",
        "hvm": "ami-0dbe2acf413e40198"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-032afb96b65bff3b9"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-06f14b5469c9df904"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0b51e8c44ea0e8fda",
        "hvm": "ami-0a118af4992024115"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0cd5233ab463d693b",
        "hvm": "ami-01c565122d2a1e4a5"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-082a215ef343a4fac"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-06ba3e4489b04eeb8",
        "hvm": "ami-05f26f4d887c6de6e"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0085060c2159a09e4"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-0447b4b8a54f9def1",
        "hvm": "ami-012be1589f2bb2c22"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-00031096d352a7e28"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0d82cf915cd1a546c",
        "hvm": "ami-0e7135de578607b68"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-090240162f36dd8ab"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0af6868b7720e388a"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0b4d3491240a2e5b8",
        "hvm": "ami-05a9e5756eceb0e31"
      },
      {
        "name": "us-east-1",
        "pv": "ami-091a5b395084c28fc",
        "hvm": "ami-0cd0b00847fe4fbca"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-0c7816bcc24a4829b"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-0a1833ef92a85903d"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-88c0b0e9",
        "hvm": "ami-fec2b29f"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0583a4eda6cb2f8bc",
        "hvm": "ami-0556214ee8ac652a7"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0ee1dd80c19fa5950",
        "hvm": "ami-09bf93c5ba7995cc8"
      }
    ]
  },
  "2079.4.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-082e32d66ec816a8d",
        "hvm": "ami-0bf35668be4576cf8"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-051e99e2756755a9e"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-0e30aa1ac88761da3"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-01fa702368c1f1ecf",
        "hvm": "ami-08b3e33226bfb93e6"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0486a4781ffc93b4f",
        "hvm": "ami-026b4e4161fd2f1cc"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-03af58b9ae2fa66d7"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-0ce4043e4929ad181",
        "hvm": "ami-09b54790f727ac576"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0cb93c9d844de0c18"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-02b169c1126b32990",
        "hvm": "ami-06142be4afc3d5b1b"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-0d1a72da6fcf503b4"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0c7de3a149508353c",
        "hvm": "ami-092743440a38c284a"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-0a7022d840891e2bf"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-057e8e6d7ade7586c"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0818037eb5efee05f",
        "hvm": "ami-0b0aa7c9a9b9e1fd7"
      },
      {
        "name": "us-east-1",
        "pv": "ami-099f85ba2b1d39738",
        "hvm": "ami-0b3b03f7480ad557c"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-0bd92699e64d9d624"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-0df81f66758e1b926"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-8391ede2",
        "hvm": "ami-388cf059"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0d972297ad73dd6c2",
        "hvm": "ami-00967fa9b13b97a7c"
      },
      {
        "name": "us-west-2",
        "pv": "ami-072111d329b2ae104",
        "hvm": "ami-0b2bce213b890b477"
      }
    ]
  },
  "2079.5.1": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-0a90481dfeb83ec03",
        "hvm": "ami-036857bdeb5b3362c"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-00f2ee0ba5c3954e9"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-060a95c11ed11c1bd"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-06ed3aad1a8df6901",
        "hvm": "ami-0da0dfdf36db6e7e1"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0ec71aab8f3ff4bc7",
        "hvm": "ami-0c245ecdf4720b5a2"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-0867c908d18a8c69e"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-00a21cf9d76d82a54",
        "hvm": "ami-0032227ab96e75a9f"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-006bc343e8c9c9b22"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-020a4a16c0f1433ca",
        "hvm": "ami-0018c6ee88479b31e"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-00e03f7974618119a"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0bf955397ef19c666",
        "hvm": "ami-018351c24af175181"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-0b91a753d4fa446b4"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0719f5491f02f1874"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0cb56b9981aa3e0f1",
        "hvm": "ami-08003539c64a2c6b9"
      },
      {
        "name": "us-east-1",
        "pv": "ami-025d42cea3f7b3588",
        "hvm": "ami-0a7247846b022222c"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-0efcfe3a15d87beff"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-0229105a89981165a"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-683c4409",
        "hvm": "ami-22384043"
      },
      {
        "name": "us-west-1",
        "pv": "ami-06190a35fb167489c",
        "hvm": "ami-062f6abca7bac0908"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0065b9ce93ac68118",
        "hvm": "ami-0e7d76904282e972b"
      }
    ]
  },
  "2079.6.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-0304ee2ddc1fb352e",
        "hvm": "ami-08515504f17364103"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-0ffa479a2c62add59"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-0535659b948bd0474"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0d913a9988f7852bf",
        "hvm": "ami-08307f4cf09de9244"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-01724c6f0a8467c15",
        "hvm": "ami-014a8902e4218bec8"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-0b3847a26fb2687af"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-06e71ba1a7d23b15d",
        "hvm": "ami-0f04713c62bec1ec2"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-00aeeb7d26c265116"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-03a50939ce4b99157",
        "hvm": "ami-0d1579b60bb706fb7"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-07cbc21d0e0668619"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0a3c97545679924b1",
        "hvm": "ami-07361abacf6d5432f"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-04a26b5fb882a2eb2"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0d55a91bafde5fa3e"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-06864e885dedeefa2",
        "hvm": "ami-02000c44222293e7f"
      },
      {
        "name": "us-east-1",
        "pv": "ami-0e9a6aa0a6e6dd3c4",
        "hvm": "ami-016a8193f03cf4a79"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-0c6f750453c5ea69b"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-08e05b47b0c2d2eb5"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-aadca4cb",
        "hvm": "ami-06d9a167"
      },
      {
        "name": "us-west-1",
        "pv": "ami-06db04fd34714a33e",
        "hvm": "ami-09425a00005361f87"
      },
      {
        "name": "us-west-2",
        "pv": "ami-097210e896ddea451",
        "hvm": "ami-023d3b5f74bd89f9b"
      }
    ]
  },
  "2135.4.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-07b83b324896e0c5a",
        "hvm": "ami-02e7b007b87514a38"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-0b5d1f638fb771cc9"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-0db4916dd31b99465"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-09f6228c13ec97c86",
        "hvm": "ami-01f2de2186e97c395"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-075b2e7180ec28e0f",
        "hvm": "ami-026d43721ef96eba8"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-07d5bae9b2c4c9df1"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-0d3c925cc78d9f309",
        "hvm": "ami-0dd65d250887524c1"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0c63b500c3173c90e"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-06a0a691e74bb9493",
        "hvm": "ami-0eb0d9bb7ad1bd1e9"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-0e3eca3c62f4c6311"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-01c6fc61e21b01708",
        "hvm": "ami-000307cf706ac9f94"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-0322cee7ff4e446ce"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-01c936a41649a8cda"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-05705f71626fa05ae",
        "hvm": "ami-0b4101a238b99a929"
      },
      {
        "name": "us-east-1",
        "pv": "ami-0dd9f1cafc09d7798",
        "hvm": "ami-00386353b49e325ba"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-064fe7e0332ae6407"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-03e5a71feb2b7afd2"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-c6286da7",
        "hvm": "ami-272d6846"
      },
      {
        "name": "us-west-1",
        "pv": "ami-07d30db97db8d5b0c",
        "hvm": "ami-070bfb410b9f148c7"
      },
      {
        "name": "us-west-2",
        "pv": "ami-017f65838a98d0078",
        "hvm": "ami-0a7e0ff8d31da1836"
      }
    ]
  },
  "2135.5.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-09673be92fa59a18c",
        "hvm": "ami-070d50353dfb032ba"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-041a583a9761bda64"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-083eec4a98ca0396b"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-06b44b502238c091d",
        "hvm": "ami-0bdf64786279efbbc"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0cc32476b1b941cf3",
        "hvm": "ami-0bb7c56044b64aa56"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-082a1a74cfc2d2403"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-0fe0dc6001c982cb6",
        "hvm": "ami-0d8ca8372e3b0aff4"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-049ed451bb483d4be"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-051b84d3e0a89fec0",
        "hvm": "ami-0cfac31dd01a5f898"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-009c476af4072d56a"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0898e2390ed497160",
        "hvm": "ami-053d1b6039e1098d4"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-09e2e4b79ea105d0f"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0a409979da233373a"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0ef096d9aa2909669",
        "hvm": "ami-0b2f9ee1da741ad19"
      },
      {
        "name": "us-east-1",
        "pv": "ami-01d492ec136ec8359",
        "hvm": "ami-02b51824b39a1d52a"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-03aa12465ead76468"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-0dc23aad3fa5a13c9"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-e35b1f82",
        "hvm": "ami-6f5d190e"
      },
      {
        "name": "us-west-1",
        "pv": "ami-084c9acb389f1801b",
        "hvm": "ami-04a1dd7b81fe80e40"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0108b87fd991ef10e",
        "hvm": "ami-071f4352a744b29aa"
      }
    ]
  },
  "2135.6.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-0e5bb4c6cf0c76942",
        "hvm": "ami-0e4285d1d637f9621"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-0b0b27a09fa29bcc0"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-05533b69d18ef0f2b"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0780842b1dbf59e17",
        "hvm": "ami-03b2848db9a1e8331"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0bd068ba067b763e5",
        "hvm": "ami-0f2a464ec2d360ab3"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-0971c6160f743d7a4"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-063500bcbb5170bff",
        "hvm": "ami-05e7c07155bf6194c"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0834edd97d31a9b8c"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-0ad74d5117746ca13",
        "hvm": "ami-034fd8c3f4026eb39"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-0eb52d157df39a702"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-06ee7238609b1dc23",
        "hvm": "ami-0b4e04c2cc22a915e"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-0ed4f5a960d2d3527"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0338e402d6f997560"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0708e15680ab23091",
        "hvm": "ami-017d7523a36c57feb"
      },
      {
        "name": "us-east-1",
        "pv": "ami-0cab4a1ff1e58707b",
        "hvm": "ami-04e51eabc8abea762"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-00893b3a357694f05"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-06cc1e14c395f91e7"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-a53370c4",
        "hvm": "ami-b5286bd4"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0137db8e321bed063",
        "hvm": "ami-00f0659e80ce3eba1"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0ee20b2a8a4304469",
        "hvm": "ami-073f5d166dc37a1bd"
      }
    ]
  },
  "2191.4.1": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-08d902effb02c9720",
        "hvm": "ami-013c107acc398a837"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-053ac41eaf4d36a28"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-00ed55139b0fac9f5"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0860b3210d93534c3",
        "hvm": "ami-0f1048a07cea508c9"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-03e207f9ec5638cfb",
        "hvm": "ami-0922e1c0bff34382a"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-04a5373222f967809"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-004fa9eba8893d340",
        "hvm": "ami-09d910867fa3db4e4"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-090a79bd4326061e3"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-023e4872cb4b75754",
        "hvm": "ami-06ce4a4316f287f65"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-036e0e582741a581a"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0fb0486510118b835",
        "hvm": "ami-0e5c332d034c5c778"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-0f5a5105bc30d5172"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-04de4c2943ebaa320"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0e2c106e57d11f914",
        "hvm": "ami-0949b7f1d534804b7"
      },
      {
        "name": "us-east-1",
        "pv": "ami-04c04c852d1244594",
        "hvm": "ami-03d5e9298ba4bf740"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-00dfc58478fdbc01e"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-0b4fdf9756b4e01a5"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-5bfdb33a",
        "hvm": "ami-89f4bae8"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0976837d4d660df35",
        "hvm": "ami-097febb11373d725b"
      },
      {
        "name": "us-west-2",
        "pv": "ami-07b962940a7950a13",
        "hvm": "ami-00156bb9f24b6ff73"
      }
    ]
  },
  "2191.5.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-004a5835f582c2a12",
        "hvm": "ami-06443443a3ad575e0"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-05385569b790d035a"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-05d7bc2359eaaecf1"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-05fbffaa8be36bea8",
        "hvm": "ami-0e69fd5ed05e58e4a"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-02a1877a910ef0350",
        "hvm": "ami-0af85d64c1d5aeae6"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-00cbc28393f9da64c"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-03e43ad940259ec49",
        "hvm": "ami-001272d09c87c54fa"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0c08167b4fb0293c1"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-0d1d20629c05be57e",
        "hvm": "ami-038cea5071a5ee580"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-01f28d71d1c924642"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-086e3b5a469eee732",
        "hvm": "ami-067301c1a68e593f5"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-0f5c4ede722171894"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-07bf54c1c2b7c368e"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-07c4c71f05ef4965a",
        "hvm": "ami-0d1ca6b44a76c404a"
      },
      {
        "name": "us-east-1",
        "pv": "ami-0f4286b76a4ebded7",
        "hvm": "ami-06d2804068b372d32"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-07ee0d30575e363c4"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-0751c20ce4cb557df"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-ae6820cf",
        "hvm": "ami-a9571fc8"
      },
      {
        "name": "us-west-1",
        "pv": "ami-095bcdd14510f1f0d",
        "hvm": "ami-0d05a67ab67139420"
      },
      {
        "name": "us-west-2",
        "pv": "ami-082f268fa33ca199b",
        "hvm": "ami-039eb9d6842534000"
      }
    ]
  },
  "2247.5.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-0782a6b605019db1d",
        "hvm": "ami-01ab077b2eef72be3"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-0a820ad558fc93aa1"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-0f24eb185e7d08350"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0e61629e92f75e800",
        "hvm": "ami-0b62b2c118a893fa6"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0d935db8cdc0b8148",
        "hvm": "ami-0660f03ea85aa5cf1"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-0732c119d7a3ebba9"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-0b60f9133647429a6",
        "hvm": "ami-0c72bded70c19de85"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0782c5d4029140db7"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-0032cf80802d3235d",
        "hvm": "ami-04ec5fe54b2e8c691"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-0494edad25b2fc83b"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-01a9cbd5cf03cfd79",
        "hvm": "ami-0a0f03879ce027f27"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-01a940c959ce4412d"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0ce903ccedd9f61b1"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-00719d8a97d67bdbc",
        "hvm": "ami-05be2c401e1260dcf"
      },
      {
        "name": "us-east-1",
        "pv": "ami-048838e93f8533bc3",
        "hvm": "ami-047242a24cd6b4bc9"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-01e0a83a996329ade"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-07edd9df1cc47bfad"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-87e7b1e6",
        "hvm": "ami-24f8ae45"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0d6da48a7a80f28c4",
        "hvm": "ami-05058d6d109b63333"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0b85411d486fabcca",
        "hvm": "ami-0044ad4295abefc56"
      }
    ]
  },
  "2247.6.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-047456540dea1ff30",
        "hvm": "ami-032635064e499aed3"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-09a8932051bd32169"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-0567b157f1a807f9b"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0f04e723e54ed0cf8",
        "hvm": "ami-00223a87a5d6bb871"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0b52deec7a432d4e6",
        "hvm": "ami-0ed11a03287b54de5"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-0727a541b88f23a95"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-0b14a46b60f8fc0c1",
        "hvm": "ami-0b9f97e1e2d5379fc"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-050a15c8007790e72"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-0c60537f510944e44",
        "hvm": "ami-0a6c83f43331c0072"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-091d1cda961041af3"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-03289e4f149e8edda",
        "hvm": "ami-045e70ebbd30d4313"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-019701cd5703f116a"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-05505c8ff9ad4cc9c"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0be5d0886db360a6c",
        "hvm": "ami-009d634c09389f89f"
      },
      {
        "name": "us-east-1",
        "pv": "ami-0b3ea3d81e68afeb4",
        "hvm": "ami-08ff99eae79739971"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-04d685bb4c2f2944b"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-013651eaf51a4dbd4"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-91f6a4f0",
        "hvm": "ami-90f6a4f1"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0578150916703a01c",
        "hvm": "ami-033bfd8ab98a2e629"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0057ccbe1c7c43ee3",
        "hvm": "ami-08e6569f682c07942"
      }
    ]
  },
  "2247.7.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-07f9c08bce2045f9a",
        "hvm": "ami-0a19c6d4270411e00"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-030af9f3f7a59da8a"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-0a3eda41a457d50ce"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0b5c1419f13e3e271",
        "hvm": "ami-076a5f17cdabba613"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0ac8cb2c671aa0574",
        "hvm": "ami-034371a7661a59b69"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-0116ae24134352961"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-040d1ec35cfc81f2e",
        "hvm": "ami-06b6c63297a7ecde9"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0c40c9c521f00671b"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-0ae4f38d6f4a5d344",
        "hvm": "ami-07e308cdb030da01e"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-05f9552ceca4ec098"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-01631bb42f45fbcb4",
        "hvm": "ami-0665a3a0f6db4753a"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-0027b296a789e6ae4"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0ef4de5d03c4694c6"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0268fa01aa689e84b",
        "hvm": "ami-0c1c5b1a3eda0b204"
      },
      {
        "name": "us-east-1",
        "pv": "ami-06c8696327d7491d4",
        "hvm": "ami-000c6a6ff12707589"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-0112970dd8ff85db4"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-01af418a7d0fb46bb"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-df5e01be",
        "hvm": "ami-5d5e013c"
      },
      {
        "name": "us-west-1",
        "pv": "ami-00e7fca288ce8be86",
        "hvm": "ami-07188d54f84835ca8"
      },
      {
        "name": "us-west-2",
        "pv": "ami-047bc48e833d214b3",
        "hvm": "ami-0de4b268ec6ba35fd"
      }
    ]
  },
  "2303.3.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-071f56a6f472a9069",
        "hvm": "ami-0a3dec216288de4d0"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-02adc372e1ee4146f"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-0bbd4b2079d5402b4"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-02e3e4ad2c16106b2",
        "hvm": "ami-01c32d64ac0ba511d"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0dde1db77b25e0b9f",
        "hvm": "ami-0419e4286d0d71faf"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-03a72248ea060789a"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-0c58deca8c21e8910",
        "hvm": "ami-074403249c8493cae"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0982d4762e829ff1b"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-05e5f9a90c9e8db16",
        "hvm": "ami-031c08681db8c400e"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-0c3ab996558c44892"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0ee47149e12bdfbd0",
        "hvm": "ami-0143712d42aa4a7c9"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-04d721db24f40ffce"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-04740abcad65f30d8"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-08ef70bbdf5ef2924",
        "hvm": "ami-03e977723db9b9ade"
      },
      {
        "name": "us-east-1",
        "pv": "ami-0c6a2ab8b533c4ad4",
        "hvm": "ami-0a953cad0391f0305"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-040ed4d275bf17303"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-07b62f136d16aca9f"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-96247df7",
        "hvm": "ami-a62079c7"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0f77fb9ece0b95cfb",
        "hvm": "ami-03a8c2f3cfe69169d"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0e8fb1f4491bb32ee",
        "hvm": "ami-0adf78a0f99af398f"
      }
    ]
  },
  "2303.4.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-04cfc008b6941c2a4",
        "hvm": "ami-0e4257472375320d1"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-0927d63c4738c56fe"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-0c5aaa83e27d00fd0"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0999bdf89931d017c",
        "hvm": "ami-0145234ddc719501c"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-04bca02e8ba8782c8",
        "hvm": "ami-0e47b8ce20b478672"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-0fd9425144ea0bfb0"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-064f3bf951d221280",
        "hvm": "ami-0c356cd59a18bdf40"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-071ebc86cd72be4af"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-05fb1ed961b6d3b98",
        "hvm": "ami-0d1523a303dd37067"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-06b16405d63171264"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0b3d107fdd43d780d",
        "hvm": "ami-0c6ca83c80e8bba91"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-0a2d34b930d813466"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-03871bbff3a643d6d"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-065e02f618741f04f",
        "hvm": "ami-0a34138b2787a9dd7"
      },
      {
        "name": "us-east-1",
        "pv": "ami-0243b3a167deacb87",
        "hvm": "ami-0f2d95e41c7dac6b4"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-0bfdb6a28829a211c"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-060989800543ae71a"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-73567b12",
        "hvm": "ami-33507d52"
      },
      {
        "name": "us-west-1",
        "pv": "ami-084ad52ff62341ffa",
        "hvm": "ami-0ea414f001b77d38b"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0657618868385c45a",
        "hvm": "ami-0c2a171c931888989"
      }
    ]
  },
  "2345.3.0": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-068e943403b24d8d1",
        "hvm": "ami-061659fcdbb942671"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-0048f1282cc2f7020"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-0576a199d1e2f2110"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-04d44bd8ec4df9c78",
        "hvm": "ami-030cef2acc6e5377f"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-083033fd2b0e1a1fb",
        "hvm": "ami-08b526947c08b5842"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-02444192766d2877f"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-0180d1fecf4988a5f",
        "hvm": "ami-026f7fae59b401ac0"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0946ef005be7e5e20"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-0536fc1831b5ff8e9",
        "hvm": "ami-06c600855f8f21e97"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-0ea6babd45136d7a6"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-098e389ba5b071943",
        "hvm": "ami-07c25af0e918ce3c1"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-06b451fafc3def0dc"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0d1154386a7a334c9"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0cc157b702cf900cb",
        "hvm": "ami-005ce0c51d9e43786"
      },
      {
        "name": "us-east-1",
        "pv": "ami-012e4bc9baf52100c",
        "hvm": "ami-07cce92cad14cc238"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-08c51fc1b1cc85501"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-024759eff71c6b4b7"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-c87f56a9",
        "hvm": "ami-c97f56a8"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0dcbc40fe33cb3678",
        "hvm": "ami-04b8d2ccf0bf3a6eb"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0e9738efa22509ffb",
        "hvm": "ami-018b1e7ac21df62b9"
      }
    ]
  },
  "494.3.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-bad3e2a7",
        "hvm": "ami-b4d3e2a9"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-5e3e3e5f",
        "hvm": "ami-603e3e61"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-8ffd4d92",
        "hvm": "ami-89fd4d94"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-6d90f957",
        "hvm": "ami-6390f959"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-c982af9b",
        "hvm": "ami-cf82af9d"
      },
      {
        "name": "us-east-1",
        "pv": "ami-8a48d0e2",
        "hvm": "ami-b648d0de"
      },
      {
        "name": "us-west-2",
        "pv": "ami-93b0e7a3",
        "hvm": "ami-91b0e7a1"
      },
      {
        "name": "us-west-1",
        "pv": "ami-a17360e4",
        "hvm": "ami-a37360e6"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-3c0cbe4b",
        "hvm": "ami-3a0cbe4d"
      }
    ]
  },
  "494.4.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-12c0f10f",
        "hvm": "ami-10c0f10d"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-dc6566dd",
        "hvm": "ami-da6566db"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-9bda6a86",
        "hvm": "ami-99da6a84"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-abc8a191",
        "hvm": "ami-a9c8a193"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-977559c5",
        "hvm": "ami-957559c7"
      },
      {
        "name": "us-east-1",
        "pv": "ami-f469f29c",
        "hvm": "ami-f669f29e"
      },
      {
        "name": "us-west-2",
        "pv": "ami-dbf8afeb",
        "hvm": "ami-d9f8afe9"
      },
      {
        "name": "us-west-1",
        "pv": "ami-af0516ea",
        "hvm": "ami-ad0516e8"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-f6853881",
        "hvm": "ami-f4853883"
      }
    ]
  },
  "494.5.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-4e7d4d53",
        "hvm": "ami-487d4d55"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-dccfc0dd",
        "hvm": "ami-decfc0df"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-c904b4d4",
        "hvm": "ami-cb04b4d6"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-d7e981ed",
        "hvm": "ami-d1e981eb"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-81406fd3",
        "hvm": "ami-83406fd1"
      },
      {
        "name": "us-east-1",
        "pv": "ami-7e5d3d16",
        "hvm": "ami-705d3d18"
      },
      {
        "name": "us-west-2",
        "pv": "ami-4fd4857f",
        "hvm": "ami-4dd4857d"
      },
      {
        "name": "us-west-1",
        "pv": "ami-15fae850",
        "hvm": "ami-17fae852"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-7a3a840d",
        "hvm": "ami-783a840f"
      }
    ]
  },
  "522.4.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-5ab78747",
        "hvm": "ami-54b78749"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-d05c4dd1",
        "hvm": "ami-d25c4dd3"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-bf3f8da2",
        "hvm": "ami-b93f8da4"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-0f8ae035",
        "hvm": "ami-0d8ae037"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-e5ae87b7",
        "hvm": "ami-ebae87b9"
      },
      {
        "name": "us-east-1",
        "pv": "ami-58562730",
        "hvm": "ami-5a562732"
      },
      {
        "name": "us-west-2",
        "pv": "ami-230c5013",
        "hvm": "ami-210c5011"
      },
      {
        "name": "us-west-1",
        "pv": "ami-a53d22e0",
        "hvm": "ami-a73d22e2"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-8d0087fa",
        "hvm": "ami-8f0087f8"
      }
    ]
  },
  "522.5.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-448dbd59",
        "hvm": "ami-468dbd5b"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-0a05160b",
        "hvm": "ami-0c05160d"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-27b00d3a",
        "hvm": "ami-23b00d3e"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-b5295c8f",
        "hvm": "ami-b7295c8d"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-ba0f27e8",
        "hvm": "ami-b40f27e6"
      },
      {
        "name": "us-east-1",
        "pv": "ami-3e750856",
        "hvm": "ami-3c750854"
      },
      {
        "name": "us-west-2",
        "pv": "ami-bf2d728f",
        "hvm": "ami-bd2d728d"
      },
      {
        "name": "us-west-1",
        "pv": "ami-8f534dca",
        "hvm": "ami-8d534dc8"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-e76dec90",
        "hvm": "ami-f96dec8e"
      }
    ]
  },
  "522.6.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-38093a25",
        "hvm": "ami-3a093a27"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-e2465fe3",
        "hvm": "ami-e4465fe5"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-7f863a62",
        "hvm": "ami-7d863a60"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-db7d09e1",
        "hvm": "ami-d97d09e3"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-3e8da66c",
        "hvm": "ami-3c8da66e"
      },
      {
        "name": "us-east-1",
        "pv": "ami-3615525e",
        "hvm": "ami-3415525c"
      },
      {
        "name": "us-west-2",
        "pv": "ami-51134b61",
        "hvm": "ami-6f134b5f"
      },
      {
        "name": "us-west-1",
        "pv": "ami-bebfa6fb",
        "hvm": "ami-bcbfa6f9"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-7bf27e0c",
        "hvm": "ami-79f27e0e"
      }
    ]
  },
  "557.2.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-88c1f295",
        "hvm": "ami-8ec1f293"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-ea5c46eb",
        "hvm": "ami-e85c46e9"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-2fe95632",
        "hvm": "ami-2de95630"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-4fd3a775",
        "hvm": "ami-4dd3a777"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-70dcf622",
        "hvm": "ami-72dcf620"
      },
      {
        "name": "us-east-1",
        "pv": "ami-8097d4e8",
        "hvm": "ami-8297d4ea"
      },
      {
        "name": "us-west-2",
        "pv": "ami-f3702bc3",
        "hvm": "ami-f1702bc1"
      },
      {
        "name": "us-west-1",
        "pv": "ami-26b5ad63",
        "hvm": "ami-24b5ad61"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-5b911f2c",
        "hvm": "ami-5d911f2a"
      }
    ]
  },
  "607.0.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-0c300d11",
        "hvm": "ami-0e300d13"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-b128dcb1",
        "hvm": "ami-af28dcaf"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-2154ec3c",
        "hvm": "ami-2354ec3e"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-bbb5c581",
        "hvm": "ami-b9b5c583"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-fa0b3aa8",
        "hvm": "ami-f80b3aaa"
      },
      {
        "name": "us-east-1",
        "pv": "ami-343b195c",
        "hvm": "ami-323b195a"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0989a439",
        "hvm": "ami-0789a437"
      },
      {
        "name": "us-west-1",
        "pv": "ami-83d533c7",
        "hvm": "ami-8dd533c9"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-57950a20",
        "hvm": "ami-55950a22"
      }
    ]
  },
  "633.1.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-8c003c91",
        "hvm": "ami-92003c8f"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-9eb9439e",
        "hvm": "ami-9cb9439c"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-19ceaf3a",
        "hvm": "ami-1bceaf38"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-9be66386",
        "hvm": "ami-99e66384"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-d53845ef",
        "hvm": "ami-cb3845f1"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-a2cefcf0",
        "hvm": "ami-a0cefcf2"
      },
      {
        "name": "us-east-1",
        "pv": "ami-d6033bbe",
        "hvm": "ami-d2033bba"
      },
      {
        "name": "us-west-2",
        "pv": "ami-39280209",
        "hvm": "ami-37280207"
      },
      {
        "name": "us-west-1",
        "pv": "ami-4df91b09",
        "hvm": "ami-43f91b07"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-2f422358",
        "hvm": "ami-21422356"
      }
    ]
  },
  "647.0.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-3c764821",
        "hvm": "ami-38764825"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-9022eb90",
        "hvm": "ami-8e22eb8e"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-cf0868ec",
        "hvm": "ami-c90868ea"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-1113940c",
        "hvm": "ami-1313940e"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-6d83fc57",
        "hvm": "ami-6383fc59"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-8a546ad8",
        "hvm": "ami-84546ad6"
      },
      {
        "name": "us-east-1",
        "pv": "ami-e8657580",
        "hvm": "ami-ea657582"
      },
      {
        "name": "us-west-2",
        "pv": "ami-67427157",
        "hvm": "ami-65427155"
      },
      {
        "name": "us-west-1",
        "pv": "ami-75dd3331",
        "hvm": "ami-77dd3333"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-4d1c763a",
        "hvm": "ami-4b1c763c"
      }
    ]
  },
  "647.2.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-fe6b52e3",
        "hvm": "ami-f86b52e5"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-9ec1119e",
        "hvm": "ami-9cc1119c"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-854828a6",
        "hvm": "ami-874828a4"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-d9b839c4",
        "hvm": "ami-d5b839c8"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-7985fc43",
        "hvm": "ami-7b85fc41"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-d4033b86",
        "hvm": "ami-da033b88"
      },
      {
        "name": "us-east-1",
        "pv": "ami-c16583aa",
        "hvm": "ami-c36583a8"
      },
      {
        "name": "us-west-2",
        "pv": "ami-995f60a9",
        "hvm": "ami-975f60a7"
      },
      {
        "name": "us-west-1",
        "pv": "ami-877a92c3",
        "hvm": "ami-857a92c1"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-e38cfc94",
        "hvm": "ami-e18cfc96"
      }
    ]
  },
  "681.0.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-a491a8b9",
        "hvm": "ami-aa91a8b7"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-a802dca8",
        "hvm": "ami-aa02dcaa"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-bfbada9c",
        "hvm": "ami-b9bada9a"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-e51797f8",
        "hvm": "ami-eb1797f6"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-b513688f",
        "hvm": "ami-ab136891"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-c0ffc592",
        "hvm": "ami-c2ffc590"
      },
      {
        "name": "us-east-1",
        "pv": "ami-7bad4710",
        "hvm": "ami-79ad4712"
      },
      {
        "name": "us-west-2",
        "pv": "ami-c7162ef7",
        "hvm": "ami-c5162ef5"
      },
      {
        "name": "us-west-1",
        "pv": "ami-f33cd6b7",
        "hvm": "ami-f13cd6b5"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-05fd8272",
        "hvm": "ami-07fd8270"
      }
    ]
  },
  "681.1.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-00f2ca1d",
        "hvm": "ami-06f2ca1b"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-e8ff5ae8",
        "hvm": "ami-eaff5aea"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-2f9cfc0c",
        "hvm": "ami-299cfc0a"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-4928ab54",
        "hvm": "ami-4b28ab56"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-f7047ecd",
        "hvm": "ami-f5047ecf"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-4c63671e",
        "hvm": "ami-4e63671c"
      },
      {
        "name": "us-east-1",
        "pv": "ami-29dc2e42",
        "hvm": "ami-2bdc2e40"
      },
      {
        "name": "us-west-2",
        "pv": "ami-f3b882c3",
        "hvm": "ami-f1b882c1"
      },
      {
        "name": "us-west-1",
        "pv": "ami-4d1aee09",
        "hvm": "ami-4f1aee0b"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-079ee570",
        "hvm": "ami-059ee572"
      }
    ]
  },
  "681.2.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-e8e5ddf5",
        "hvm": "ami-eae5ddf7"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-1c6fca1c",
        "hvm": "ami-1a6fca1a"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-ef9fffcc",
        "hvm": "ami-e99fffca"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-b3cb49ae",
        "hvm": "ami-b1cb49ac"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-2d641e17",
        "hvm": "ami-23641e19"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-d803078a",
        "hvm": "ami-da030788"
      },
      {
        "name": "us-east-1",
        "pv": "ami-91ea17fa",
        "hvm": "ami-93ea17f8"
      },
      {
        "name": "us-west-2",
        "pv": "ami-5f4d486f",
        "hvm": "ami-5d4d486d"
      },
      {
        "name": "us-west-1",
        "pv": "ami-cb67938f",
        "hvm": "ami-c967938d"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-512f5526",
        "hvm": "ami-5f2f5528"
      }
    ]
  },
  "717.1.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-8e764c93",
        "hvm": "ami-8c764c91"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-ca8926ca",
        "hvm": "ami-cc8926cc"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-812a49a2",
        "hvm": "ami-832a49a0"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-63cf437e",
        "hvm": "ami-9dcf4380"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-69a1e553",
        "hvm": "ami-6fa1e555"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-ee6b6bbc",
        "hvm": "ami-e86b6bba"
      },
      {
        "name": "us-east-1",
        "pv": "ami-9d5894f6",
        "hvm": "ami-9f5894f4"
      },
      {
        "name": "us-west-2",
        "pv": "ami-cf8a8bff",
        "hvm": "ami-cd8a8bfd"
      },
      {
        "name": "us-west-1",
        "pv": "ami-4d27d709",
        "hvm": "ami-4f27d70b"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-22410355",
        "hvm": "ami-20410357"
      }
    ]
  },
  "717.3.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-00211b1d",
        "hvm": "ami-02211b1f"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-24d27b24",
        "hvm": "ami-22d27b22"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-eb3a59c8",
        "hvm": "ami-e53a59c6"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-47a62a5a",
        "hvm": "ami-45a62a58"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-292e6913",
        "hvm": "ami-2b2e6911"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0cf1f15e",
        "hvm": "ami-0ef1f15c"
      },
      {
        "name": "us-east-1",
        "pv": "ami-691cd402",
        "hvm": "ami-6b1cd400"
      },
      {
        "name": "us-west-2",
        "pv": "ami-f7a5a5c7",
        "hvm": "ami-f5a5a5c5"
      },
      {
        "name": "us-west-1",
        "pv": "ami-bd8477f9",
        "hvm": "ami-bf8477fb"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-52f4b925",
        "hvm": "ami-50f4b927"
      }
    ]
  },
  "723.3.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-b8cecaa5",
        "hvm": "ami-bececaa3"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-f0338ff0",
        "hvm": "ami-f2338ff2"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-c55033e6",
        "hvm": "ami-c75033e4"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-17e9600a",
        "hvm": "ami-11e9600c"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-8988c8b3",
        "hvm": "ami-8f88c8b5"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-b4d8d4e6",
        "hvm": "ami-b6d8d4e4"
      },
      {
        "name": "us-east-1",
        "pv": "ami-3b73d350",
        "hvm": "ami-3d73d356"
      },
      {
        "name": "us-west-2",
        "pv": "ami-87ada4b7",
        "hvm": "ami-85ada4b5"
      },
      {
        "name": "us-west-1",
        "pv": "ami-1fb04f5b",
        "hvm": "ami-1db04f59"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-0c10417b",
        "hvm": "ami-0e104179"
      }
    ]
  },
  "766.3.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-76bbba6b",
        "hvm": "ami-74bbba69"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-2277ff22",
        "hvm": "ami-1e77ff1e"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-f7d1b2d4",
        "hvm": "ami-f1d1b2d2"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-9d2ba180",
        "hvm": "ami-632ba17e"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-8df8b4b7",
        "hvm": "ami-83f8b4b9"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-10060c42",
        "hvm": "ami-12060c40"
      },
      {
        "name": "us-east-1",
        "pv": "ami-fd96fa98",
        "hvm": "ami-f396fa96"
      },
      {
        "name": "us-west-2",
        "pv": "ami-9bbfadab",
        "hvm": "ami-99bfada9"
      },
      {
        "name": "us-west-1",
        "pv": "ami-e5e71da1",
        "hvm": "ami-dbe71d9f"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-85e9c8f2",
        "hvm": "ami-83e9c8f4"
      }
    ]
  },
  "766.4.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-860a089b",
        "hvm": "ami-840a0899"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-6a5ac56a",
        "hvm": "ami-6c5ac56c"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-7f6a085c",
        "hvm": "ami-796a085a"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-2d960130",
        "hvm": "ami-3396012e"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-ebace5d1",
        "hvm": "ami-f5ace5cf"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-40819612",
        "hvm": "ami-46819614"
      },
      {
        "name": "us-east-1",
        "pv": "ami-07783d62",
        "hvm": "ami-05783d60"
      },
      {
        "name": "us-west-2",
        "pv": "ami-ef8b90df",
        "hvm": "ami-ed8b90dd"
      },
      {
        "name": "us-west-1",
        "pv": "ami-2929ee6d",
        "hvm": "ami-2b29ee6f"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-e997bc9e",
        "hvm": "ami-eb97bc9c"
      }
    ]
  },
  "766.5.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-202c3f4c",
        "hvm": "ami-fdd4c791"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-cceec9a2",
        "hvm": "ami-84e0c7ea"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-8ebc01ef",
        "hvm": "ami-05bc0164"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-bc48f3d0",
        "hvm": "ami-154af179"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-d44a14b7",
        "hvm": "ami-f35b0590"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-0065a263",
        "hvm": "ami-da67a0b9"
      },
      {
        "name": "us-east-1",
        "pv": "ami-68bdc102",
        "hvm": "ami-37bdc15d"
      },
      {
        "name": "us-west-2",
        "pv": "ami-a2ebfcc3",
        "hvm": "ami-00ebfc61"
      },
      {
        "name": "us-west-1",
        "pv": "ami-38533c58",
        "hvm": "ami-27553a47"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-05d30a76",
        "hvm": "ami-55d20b26"
      }
    ]
  },
  "835.10.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-d62d34ba",
        "hvm": "ami-3f2c3553"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-579ca339",
        "hvm": "ami-4b9fa025"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-4a02bf2b",
        "hvm": "ami-4cfd412d"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-871091eb",
        "hvm": "ami-45149529"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-d5dbfeb6",
        "hvm": "ami-14a58077"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-3a37fb59",
        "hvm": "ami-ea3af689"
      },
      {
        "name": "us-east-1",
        "pv": "ami-4057712a",
        "hvm": "ami-ee527484"
      },
      {
        "name": "us-west-2",
        "pv": "ami-3c6c895c",
        "hvm": "ami-1b61847b"
      },
      {
        "name": "us-west-1",
        "pv": "ami-c3a4d0a3",
        "hvm": "ami-b0a6d2d0"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-4412b837",
        "hvm": "ami-581db72b"
      }
    ]
  },
  "835.11.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-02e7fe6e",
        "hvm": "ami-fee2fb92"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-4b043a25",
        "hvm": "ami-26033d48"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-4df8442c",
        "hvm": "ami-bdf04cdc"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-99c444f5",
        "hvm": "ami-10c5457c"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-dc88adbf",
        "hvm": "ami-dc8baebf"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-046fa367",
        "hvm": "ami-6969a50a"
      },
      {
        "name": "us-east-1",
        "pv": "ami-1422037e",
        "hvm": "ami-23260749"
      },
      {
        "name": "us-west-2",
        "pv": "ami-c09276a0",
        "hvm": "ami-20927640"
      },
      {
        "name": "us-west-1",
        "pv": "ami-e8e59188",
        "hvm": "ami-c2e490a2"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-c377c2b0",
        "hvm": "ami-7e72c70d"
      }
    ]
  },
  "835.12.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-e8ebf384",
        "hvm": "ami-f0e8f09c"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-673b0109",
        "hvm": "ami-a93802c7"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-87d66ae6",
        "hvm": "ami-46e05c27"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-a41b9bc8",
        "hvm": "ami-6c1a9a00"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-2da1854e",
        "hvm": "ami-d0a783b3"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-8f69a6ec",
        "hvm": "ami-4a65aa29"
      },
      {
        "name": "us-east-1",
        "pv": "ami-94b49bfe",
        "hvm": "ami-dfb699b5"
      },
      {
        "name": "us-west-2",
        "pv": "ami-e6c82e86",
        "hvm": "ami-abc82ecb"
      },
      {
        "name": "us-west-1",
        "pv": "ami-912d5bf1",
        "hvm": "ami-4d2d5b2d"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-9063d5e3",
        "hvm": "ami-1461d767"
      }
    ]
  },
  "835.13.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-45190329",
        "hvm": "ami-15190379"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-7dc4c513",
        "hvm": "ami-02c9c86c"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-94b10df5",
        "hvm": "ami-e0b70b81"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-c70784ab",
        "hvm": "ami-c40784a8"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-569fb835",
        "hvm": "ami-949abdf7"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-5da46d3e",
        "hvm": "ami-00a06963"
      },
      {
        "name": "us-east-1",
        "pv": "ami-ec3d0c86",
        "hvm": "ami-7f3a0b15"
      },
      {
        "name": "us-west-2",
        "pv": "ami-7c01e21c",
        "hvm": "ami-4f00e32f"
      },
      {
        "name": "us-west-1",
        "pv": "ami-6793e207",
        "hvm": "ami-a8aedfc8"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-4b18aa38",
        "hvm": "ami-2a1fad59"
      }
    ]
  },
  "835.8.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-8a2c31e6",
        "hvm": "ami-9f2f32f3"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-96e4cbf8",
        "hvm": "ami-cde8c7a3"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-64962b05",
        "hvm": "ami-57972a36"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-497aff25",
        "hvm": "ami-487aff24"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-b7237bd4",
        "hvm": "ami-5c267e3f"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-f8c6069b",
        "hvm": "ami-6ec7070d"
      },
      {
        "name": "us-east-1",
        "pv": "ami-086a2862",
        "hvm": "ami-1a642670"
      },
      {
        "name": "us-west-2",
        "pv": "ami-5e3a283f",
        "hvm": "ami-e0342681"
      },
      {
        "name": "us-west-1",
        "pv": "ami-eec4ad8e",
        "hvm": "ami-23c0a943"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-da06a1a9",
        "hvm": "ami-5202a521"
      }
    ]
  },
  "835.9.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-72b2af1e",
        "hvm": "ami-ffafb293"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-06eac368",
        "hvm": "ami-dae8c1b4"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-a88934c9",
        "hvm": "ami-a98e33c8"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-a89f1bc4",
        "hvm": "ami-4e981c22"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-ffaef69c",
        "hvm": "ami-eeadf58d"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-ef57978c",
        "hvm": "ami-085a9a6b"
      },
      {
        "name": "us-east-1",
        "pv": "ami-05ffb06f",
        "hvm": "ami-cbfdb2a1"
      },
      {
        "name": "us-west-2",
        "pv": "ami-35c8d554",
        "hvm": "ami-16cfd277"
      },
      {
        "name": "us-west-1",
        "pv": "ami-e9aec689",
        "hvm": "ami-0eacc46e"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-fd6ccd8e",
        "hvm": "ami-c26bcab1"
      }
    ]
  },
  "899.13.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-e18c6a8e",
        "hvm": "ami-cb8d6ba4"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-b62336d8",
        "hvm": "ami-962c39f8"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-2841fd49",
        "hvm": "ami-0f3c806e"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-92a26bfc",
        "hvm": "ami-03a76e6d"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-2c961a40",
        "hvm": "ami-a49915c8"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-92dffff1",
        "hvm": "ami-74dcfc17"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-e58e4486",
        "hvm": "ami-3b8f4558"
      },
      {
        "name": "us-east-1",
        "pv": "ami-1c3c3076",
        "hvm": "ami-2c393546"
      },
      {
        "name": "us-west-2",
        "pv": "ami-57779f37",
        "hvm": "ami-4f4ba32f"
      },
      {
        "name": "us-west-1",
        "pv": "ami-7ce4961c",
        "hvm": "ami-52e69432"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-a344c0d0",
        "hvm": "ami-c346c2b0"
      }
    ]
  },
  "899.15.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-713fde1e",
        "hvm": "ami-e13fde8e"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-7504141b",
        "hvm": "ami-e304148d"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-f119a590",
        "hvm": "ami-cf19a5ae"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-3d1bd253",
        "hvm": "ami-131dd47d"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-b543ccd9",
        "hvm": "ami-d75bd4bb"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-ef84a78c",
        "hvm": "ami-a184a7c2"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-9ebb6efd",
        "hvm": "ami-52a07531"
      },
      {
        "name": "us-east-1",
        "pv": "ami-fa627590",
        "hvm": "ami-7a627510"
      },
      {
        "name": "us-west-2",
        "pv": "ami-8c7c89ec",
        "hvm": "ami-4f7f8a2f"
      },
      {
        "name": "us-west-1",
        "pv": "ami-ff75099f",
        "hvm": "ami-d8770bb8"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-51931322",
        "hvm": "ami-3b941448"
      }
    ]
  },
  "899.17.0": {
    "amis": [
      {
        "name": "eu-central-1",
        "pv": "ami-ea1cfe85",
        "hvm": "ami-021ffd6d"
      },
      {
        "name": "ap-northeast-1",
        "pv": "ami-0efce660",
        "hvm": "ami-b4fce6da"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-149d2275",
        "hvm": "ami-519c2330"
      },
      {
        "name": "ap-northeast-2",
        "pv": "ami-e04b838e",
        "hvm": "ami-d14d85bf"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-36bf365a",
        "hvm": "ami-294ec745"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-fc527e9f",
        "hvm": "ami-a32c00c0"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-ae02d5cd",
        "hvm": "ami-b401d6d7"
      },
      {
        "name": "us-east-1",
        "pv": "ami-2266874f",
        "hvm": "ami-8d6485e0"
      },
      {
        "name": "us-west-2",
        "pv": "ami-31b94b51",
        "hvm": "ami-f5bc4e95"
      },
      {
        "name": "us-west-1",
        "pv": "ami-152d5475",
        "hvm": "ami-652c5505"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-f0ab2383",
        "hvm": "ami-feaa228d"
      }
    ]
  },
  "current": {
    "amis": [
      {
        "name": "ap-northeast-1",
        "pv": "ami-068e943403b24d8d1",
        "hvm": "ami-061659fcdbb942671"
      },
      {
        "name": "ap-northeast-2",
        "pv": "",
        "hvm": "ami-0048f1282cc2f7020"
      },
      {
        "name": "ap-south-1",
        "pv": "",
        "hvm": "ami-0576a199d1e2f2110"
      },
      {
        "name": "ap-southeast-1",
        "pv": "ami-04d44bd8ec4df9c78",
        "hvm": "ami-030cef2acc6e5377f"
      },
      {
        "name": "ap-southeast-2",
        "pv": "ami-083033fd2b0e1a1fb",
        "hvm": "ami-08b526947c08b5842"
      },
      {
        "name": "ca-central-1",
        "pv": "",
        "hvm": "ami-02444192766d2877f"
      },
      {
        "name": "cn-north-1",
        "pv": "ami-0180d1fecf4988a5f",
        "hvm": "ami-026f7fae59b401ac0"
      },
      {
        "name": "cn-northwest-1",
        "pv": "",
        "hvm": "ami-0946ef005be7e5e20"
      },
      {
        "name": "eu-central-1",
        "pv": "ami-0536fc1831b5ff8e9",
        "hvm": "ami-06c600855f8f21e97"
      },
      {
        "name": "eu-north-1",
        "pv": "",
        "hvm": "ami-0ea6babd45136d7a6"
      },
      {
        "name": "eu-west-1",
        "pv": "ami-098e389ba5b071943",
        "hvm": "ami-07c25af0e918ce3c1"
      },
      {
        "name": "eu-west-2",
        "pv": "",
        "hvm": "ami-06b451fafc3def0dc"
      },
      {
        "name": "eu-west-3",
        "pv": "",
        "hvm": "ami-0d1154386a7a334c9"
      },
      {
        "name": "sa-east-1",
        "pv": "ami-0cc157b702cf900cb",
        "hvm": "ami-005ce0c51d9e43786"
      },
      {
        "name": "us-east-1",
        "pv": "ami-012e4bc9baf52100c",
        "hvm": "ami-07cce92cad14cc238"
      },
      {
        "name": "us-east-2",
        "pv": "",
        "hvm": "ami-08c51fc1b1cc85501"
      },
      {
        "name": "us-gov-east-1",
        "pv": "",
        "hvm": "ami-024759eff71c6b4b7"
      },
      {
        "name": "us-gov-west-1",
        "pv": "ami-c87f56a9",
        "hvm": "ami-c97f56a8"
      },
      {
        "name": "us-west-1",
        "pv": "ami-0dcbc40fe33cb3678",
        "hvm": "ami-04b8d2ccf0bf3a6eb"
      },
      {
        "name": "us-west-2",
        "pv": "ami-0e9738efa22509ffb",
        "hvm": "ami-018b1e7ac21df62b9"
      }
    ]
  }
}`)
var amiInfo = map[string]AMIInfoList{}

func init() {
	err := json.Unmarshal(amiJSON, &amiInfo)
	if err != nil {
		panic(err)
	}
}

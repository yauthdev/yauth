import React, { useEffect } from "react";
import {
  Box,
  Divider,
  Flex,
  Stack,
  Center,
  Text,
  Button,
  Input,
  InputGroup,
  InputRightElement,
  useToast,
} from "@chakra-ui/react";
import { useClient } from "urql";
import { FaSave } from "react-icons/fa";
import _ from "lodash";
import { EnvVariablesQuery } from "../graphql/queries";
import {
  ArrayInputType,
  SelectInputType,
  HiddenInputType,
  TextInputType,
  TextAreaInputType,
  SwitchInputType,
  HMACEncryptionType,
  RSAEncryptionType,
  ECDSAEncryptionType,
  envVarTypes,
} from "../constants";
import { UpdateEnvVariables } from "../graphql/mutation";
import { getObjectDiff, capitalizeFirstLetter } from "../utils";
// import GenerateKeysModal from "../components/GenerateKeysModal";

// Component inputs
import InputField from "../components/InputField";
import InstanceInformation from "../components/EnvComponents/InstanceInformation/InstanceInformation";
import SocialMediaLogin from "../components/EnvComponents/SocialMediaLogin/SocialMediaLogin";
import Roles from "../components/EnvComponents/Roles/Roles";
import JWTConfigurations from "../components/EnvComponents/JWTConfigurations/JSTConfigurations";
import SessionStorage from "../components/EnvComponents/SessionStorage/SessionStorage";
import EmailConfigurations from "../components/EnvComponents/EmailConfigurations/EmailConfigurations";
import WhiteListing from "../components/EnvComponents/WhiteListing/WhiteListing";
import OrganizationInfo from "../components/EnvComponents/OrganizationInfo/OrganizationInfo";
import AccessToken from "../components/EnvComponents/AccessToken/AccessToken";
import DisableFeature from "../components/EnvComponents/DisableFeature/DisableFeature";
import DangerArea from "../components/EnvComponents/DangerArea/DangerArea";

export default function Environment() {
  const client = useClient();
  const toast = useToast();
  const [adminSecret, setAdminSecret] = React.useState<
    Record<string, string | boolean>
  >({
    value: "",
    disableInputField: true,
  });
  const [loading, setLoading] = React.useState<boolean>(true);
  const [envVariables, setEnvVariables] = React.useState<envVarTypes>({
    GOOGLE_CLIENT_ID: "",
    GOOGLE_CLIENT_SECRET: "",
    GITHUB_CLIENT_ID: "",
    GITHUB_CLIENT_SECRET: "",
    FACEBOOK_CLIENT_ID: "",
    FACEBOOK_CLIENT_SECRET: "",
    ROLES: [],
    DEFAULT_ROLES: [],
    PROTECTED_ROLES: [],
    JWT_TYPE: "",
    JWT_SECRET: "",
    JWT_ROLE_CLAIM: "",
    JWT_PRIVATE_KEY: "",
    JWT_PUBLIC_KEY: "",
    REDIS_URL: "",
    SMTP_HOST: "",
    SMTP_PORT: "",
    SMTP_USERNAME: "",
    SMTP_PASSWORD: "",
    SENDER_EMAIL: "",
    ALLOWED_ORIGINS: [],
    ORGANIZATION_NAME: "",
    ORGANIZATION_LOGO: "",
    CUSTOM_ACCESS_TOKEN_SCRIPT: "",
    ADMIN_SECRET: "",
    DISABLE_LOGIN_PAGE: false,
    DISABLE_MAGIC_LINK_LOGIN: false,
    DISABLE_EMAIL_VERIFICATION: false,
    DISABLE_BASIC_AUTHENTICATION: false,
    DISABLE_SIGN_UP: false,
    OLD_ADMIN_SECRET: "",
    DATABASE_NAME: "",
    DATABASE_TYPE: "",
    DATABASE_URL: "",
    ACCESS_TOKEN_EXPIRY_TIME: "",
  });

  const [fieldVisibility, setFieldVisibility] = React.useState<
    Record<string, boolean>
  >({
    GOOGLE_CLIENT_SECRET: false,
    GITHUB_CLIENT_SECRET: false,
    FACEBOOK_CLIENT_SECRET: false,
    JWT_SECRET: false,
    SMTP_PASSWORD: false,
    ADMIN_SECRET: false,
    OLD_ADMIN_SECRET: false,
  });

  async function getData() {
    const {
      data: { _env: envData },
    } = await client.query(EnvVariablesQuery).toPromise();
    setLoading(false);
    setEnvVariables({
      ...envData,
      OLD_ADMIN_SECRET: envData.ADMIN_SECRET,
      ADMIN_SECRET: "",
    });
    setAdminSecret({
      value: "",
      disableInputField: true,
    });
  }

  useEffect(() => {
    getData();
  }, []);

  const validateAdminSecretHandler = (event: any) => {
    if (envVariables.OLD_ADMIN_SECRET === event.target.value) {
      setAdminSecret({
        ...adminSecret,
        value: event.target.value,
        disableInputField: false,
      });
    } else {
      setAdminSecret({
        ...adminSecret,
        value: event.target.value,
        disableInputField: true,
      });
    }
    if (envVariables.ADMIN_SECRET !== "") {
      setEnvVariables({ ...envVariables, ADMIN_SECRET: "" });
    }
  };

  const saveHandler = async () => {
    setLoading(true);
    const {
      data: { _env: envData },
    } = await client.query(EnvVariablesQuery).toPromise();
    const diff = getObjectDiff(envVariables, envData);
    const updatedEnvVariables = diff.reduce(
      (acc: any, property: string) => ({
        ...acc,
        // @ts-ignore
        [property]: envVariables[property],
      }),
      {}
    );
    if (
      updatedEnvVariables[HiddenInputType.ADMIN_SECRET] === "" ||
      updatedEnvVariables[HiddenInputType.OLD_ADMIN_SECRET] !==
        envData.ADMIN_SECRET
    ) {
      delete updatedEnvVariables.OLD_ADMIN_SECRET;
      delete updatedEnvVariables.ADMIN_SECRET;
    }

    delete updatedEnvVariables.DATABASE_URL;
    delete updatedEnvVariables.DATABASE_TYPE;
    delete updatedEnvVariables.DATABASE_NAME;

    const res = await client
      .mutation(UpdateEnvVariables, { params: updatedEnvVariables })
      .toPromise();

    setLoading(false);

    if (res.error) {
      toast({
        title: capitalizeFirstLetter(res.error.message),
        isClosable: true,
        status: "error",
        position: "bottom-right",
      });

      return;
    }

    setAdminSecret({
      value: "",
      disableInputField: true,
    });

    getData();

    toast({
      title: `Successfully updated ${
        Object.keys(updatedEnvVariables).length
      } variables`,
      isClosable: true,
      status: "success",
      position: "bottom-right",
    });
  };

  return (
    <Box m="5" py="5" px="10" bg="white" rounded="md">
      <InstanceInformation
        envVariables={envVariables}
        setVariables={setEnvVariables}
        inputType={TextInputType.CLIENT_ID}
        readOnly={true}
        fieldVisibility={fieldVisibility}
        setFieldVisibility={setFieldVisibility}
      />
      <Divider marginTop="2%" marginBottom="2%" />
      <SocialMediaLogin
        variables={envVariables}
        setVariables={setEnvVariables}
        fieldVisibility={fieldVisibility}
        setFieldVisibility={setFieldVisibility}
      />
      <Divider marginTop="2%" marginBottom="2%" />
      <Roles variables={envVariables} setVariables={setEnvVariables} />
      <Divider marginTop="2%" marginBottom="2%" />
      <JWTConfigurations
        variables={envVariables}
        setVariables={setEnvVariables}
        fieldVisibility={fieldVisibility}
        setFieldVisibility={setFieldVisibility}
        SelectInputType={SelectInputType.JWT_TYPE}
        // value={SelectInputType.JWT_TYPE}
        HMACEncryptionType={HMACEncryptionType}
        RSAEncryptionType={RSAEncryptionType}
        ECDSAEncryptionType={ECDSAEncryptionType}
        getData={getData}
      />
      <Divider marginTop="2%" marginBottom="2%" />
      <SessionStorage
        variables={envVariables}
        setVariables={setEnvVariables}
        RedisURL={TextInputType.REDIS_URL}
      />
      <Divider marginTop="2%" marginBottom="2%" />

      <EmailConfigurations
        variables={envVariables}
        setVariables={setEnvVariables}
        fieldVisibility={fieldVisibility}
        setFieldVisibility={setFieldVisibility}
      />
      <Divider marginTop="2%" marginBottom="2%" />

      <WhiteListing variables={envVariables} setVariables={setEnvVariables} />
      <Divider marginTop="2%" marginBottom="2%" />

      <OrganizationInfo
        variables={envVariables}
        setVariables={setEnvVariables}
      />
      <Divider marginTop="2%" marginBottom="2%" />

      <AccessToken variables={envVariables} setVariables={setEnvVariables} />
      <Divider marginTop="2%" marginBottom="2%" />

      <DisableFeature variables={envVariables} setVariables={setEnvVariables} />
      <Divider marginTop="2%" marginBottom="2%" />

      <DangerArea
        variables={envVariables}
        setVariables={setEnvVariables}
        fieldVisibility={fieldVisibility}
        setFieldVisibility={setFieldVisibility}
        validateAdminSecretHandler={validateAdminSecretHandler}
        adminSecret={adminSecret}
      />
      <Divider marginTop="5%" marginBottom="2%" />
      <Stack spacing={6} padding="1% 0">
        <Flex justifyContent="end" alignItems="center">
          <Button
            leftIcon={<FaSave />}
            colorScheme="blue"
            variant="solid"
            onClick={saveHandler}
            isDisabled={loading}
          >
            Save
          </Button>
        </Flex>
      </Stack>
    </Box>
  );
}

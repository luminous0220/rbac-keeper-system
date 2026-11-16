import React from "react";
import { Button, Result } from "antd";
import { useNavigate } from "react-router";
import { CFG } from "@/constants";
import { useTranslation } from "react-i18next";

const NotFound: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();

  return (
    <Result
      status="404"
      title="404"
      subTitle={t("notFound.subtitle")}
      extra={
        <Button type="primary" size="large" onClick={() => navigate(CFG.HOME_PATH)}>
          {t("notFound.home")}
        </Button>
      }
    />
  );
};

export default NotFound;

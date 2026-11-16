import {
  IconLayoutSidebar,
  IconLayoutSidebarRight,
} from "@tabler/icons-react";
import { Button } from "antd";
import { useThemeStore } from "@/stores/theme";
import { useTranslation } from "react-i18next";

export interface CollapseProps {
  className?: string;
  collapsed?: boolean;
  setCollapsed?: (collapsed: boolean) => void;
}

export const CollapseButton = ({
  collapsed,
  setCollapsed,
  className = "",
}: CollapseProps) => {
  const { colorVariants } = useThemeStore();
  const { t } = useTranslation();

  const toggleCollapsed = () => {
    const newCollapsed = !collapsed;
    setCollapsed?.(newCollapsed);
  };

  return (
    <Button
      className={className}
      type="text"
      icon={
        collapsed ? (
          <IconLayoutSidebar className="size-5  " stroke={2} />
        ) : (
          <IconLayoutSidebarRight className="size-5" stroke={2} />
        )
      }
      onClick={toggleCollapsed}
      aria-label={collapsed ? t("layout.expandMenu") : t("layout.collapseMenu")}
      style={{
        transition: "all 0.2s",
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.background = colorVariants.p100;
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.background = "transparent";
      }}
    />
  );
};

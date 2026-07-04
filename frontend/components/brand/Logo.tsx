import { cn } from "@/lib/utils";

type LogoProps = {
  className?: string;
  size?: "sm" | "md" | "lg" | "hero";
};

const sizeMap = {
  sm: "h-8 w-8",
  md: "h-10 w-10",
  lg: "h-14 w-14",
  hero: "h-28 w-28 md:h-36 md:w-36",
};

export function Logo({ className, size = "md" }: LogoProps) {
  return (
    <img
      src="/logo.png"
      alt="The Idea Guy"
      className={cn("object-contain", sizeMap[size], className)}
    />
  );
}
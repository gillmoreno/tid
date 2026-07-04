import { cn } from "@/lib/utils";

type LogoProps = {
  className?: string;
  size?: "sm" | "md" | "lg" | "hero";
};

const sizeMap = {
  sm: "h-8 w-auto max-w-[4.5rem]",
  md: "h-10 w-auto max-w-[5.5rem]",
  lg: "h-14 w-auto max-w-[8rem]",
  hero: "h-28 w-auto max-w-[18rem] md:h-36 md:max-w-[22rem]",
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
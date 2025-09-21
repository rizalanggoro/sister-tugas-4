import { cn } from "@/lib/utils";
import { Card } from "./ui/card";

interface MessageBubbleProps {
  name: string;
  message: string;
  position: "left" | "right";
}
export const MessageBubble = ({
  name,
  message,
  position,
}: MessageBubbleProps) => {
  const isRight = position === "right";

  return (
    <>
      <div className={cn("flex", isRight && "justify-end")}>
        <Card
          className={cn(
            "px-4 py-2 shadow-none w-fit max-w-9/12",
            isRight && "bg-primary"
          )}
        >
          <div>
            {isRight || (
              <p
                className={cn(
                  "text-xs font-medium",
                  isRight && "text-primary-foreground"
                )}
              >
                {name}
              </p>
            )}
            <p
              className={cn(
                "text-muted-foreground",
                isRight && "text-primary-foreground/70"
              )}
            >
              {message}
            </p>
          </div>
        </Card>
      </div>
    </>
  );
};

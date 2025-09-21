import { useQuery } from "@tanstack/react-query";
import { LoaderIcon, UserIcon } from "lucide-react";
import { useEffect, useState } from "react";
import ScrollToBottom from "react-scroll-to-bottom";
import useWebSocket from "react-use-websocket";
import { MessageBubble, UpdateNameDialog } from "./components";
import { CreateMessage } from "./components/create-message";
import { Button } from "./components/ui/button";

interface Chat {
  name: string;
  message: string;
}

export const App = () => {
  const [messages, setMessages] = useState<Chat[]>([]);
  const [isUpdateNameOpen, setIsUpdateNameOpen] = useState(false);
  const { lastJsonMessage } = useWebSocket(
    `${import.meta.env.VITE_WS_BASE_URL}/ws`
  );

  const name = localStorage.getItem("global-chat-name");

  const { isLoading, isSuccess, data } = useQuery({
    queryKey: ["global-messages"],
    queryFn: () =>
      fetch(`${import.meta.env.VITE_API_BASE_URL}/global-messages`).then(
        (res) => res.json()
      ),
  });

  useEffect(() => {
    if (data) {
      const { messages } = data;
      setMessages(messages);
    }
  }, [data]);

  useEffect(() => {
    try {
      if (lastJsonMessage) {
        setMessages((prev) => [...prev, lastJsonMessage as Chat]);
      }
    } catch (e) {
      console.log(e);
    }
  }, [lastJsonMessage]);

  return (
    <>
      <div className="bg-muted h-screen w-full">
        <div className="max-w-lg mx-auto h-screen flex flex-col bg-background">
          <div className="h-16 flex items-center border-b px-4 justify-between">
            <div>
              <p className="font-medium">Global Chat</p>
              <p className="text-sm text-muted-foreground">
                {name ?? "nama belum disetel"}
              </p>
            </div>

            <Button
              size={"icon"}
              variant={"outline"}
              onClick={() => setIsUpdateNameOpen(true)}
            >
              <UserIcon />
            </Button>
          </div>

          <ScrollToBottom className="flex-1 overflow-y-hidden">
            <div className="space-y-1 px-4 py-6">
              {/* loading state */}
              {isLoading && (
                <div className="flex items-center p-6 justify-center">
                  <LoaderIcon className="animate-spin" />
                </div>
              )}

              {/* success state */}
              {isSuccess &&
                messages.map((item, index) => (
                  <MessageBubble
                    key={"bubble-chat-item-" + index}
                    name={item.name}
                    message={item.message}
                    position={name === item.name ? "right" : "left"}
                  />
                ))}
            </div>
          </ScrollToBottom>

          {name && <CreateMessage />}
        </div>
      </div>

      {/* dialogs */}
      <UpdateNameDialog
        open={isUpdateNameOpen}
        onOpenChange={setIsUpdateNameOpen}
      />
    </>
  );
};

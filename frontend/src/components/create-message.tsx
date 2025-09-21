import { zodResolver } from "@hookform/resolvers/zod/dist/zod.js";
import { useMutation } from "@tanstack/react-query";
import { LoaderIcon, SendIcon } from "lucide-react";
import { useForm } from "react-hook-form";
import z from "zod";
import { Button } from "./ui/button";
import { Form, FormControl, FormField, FormItem, FormMessage } from "./ui/form";
import { Input } from "./ui/input";

const formSchema = z.object({
  message: z
    .string("Pesan tidak boleh kosong!")
    .min(1, "Pesan tidak boleh kosong!"),
});

interface Chat {
  name: string;
  message: string;
}

export const CreateMessage = () => {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
  });

  const { mutate, isPending } = useMutation({
    mutationFn: async (data: Chat) => {
      fetch(`${import.meta.env.VITE_API_BASE_URL}/global-messages`, {
        method: "POST",
        body: JSON.stringify(data),
      });
    },
    onSuccess: () => {
      form.reset({ message: "" });
    },
  });

  const onSubmit = (data: z.infer<typeof formSchema>) => {
    const name = localStorage.getItem("global-chat-name");
    if (name) mutate({ name, message: data.message });
  };

  return (
    <>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)}>
          <FormField
            control={form.control}
            name="message"
            render={({ field }) => (
              <FormItem>
                <div className="flex items-center gap-2 p-4">
                  <FormControl>
                    <Input placeholder="Masukkan pesan" {...field} />
                  </FormControl>
                  <Button type="submit" size={"icon"} disabled={isPending}>
                    {isPending ? (
                      <LoaderIcon className="animate-spin" />
                    ) : (
                      <SendIcon />
                    )}
                  </Button>
                </div>
                <FormMessage />
              </FormItem>
            )}
          />
        </form>
      </Form>
    </>
  );
};

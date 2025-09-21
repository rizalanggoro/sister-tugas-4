import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import z from "zod";
import { Button } from "./ui/button";

const formSchema = z.object({
  name: z.string("Nama tidak boleh kosong!").min(3, "Nama tidak boleh kosong!"),
});

interface UpdateNameDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}
export const UpdateNameDialog = ({
  open,
  onOpenChange,
}: UpdateNameDialogProps) => {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
  });

  useEffect(() => {
    const name = localStorage.getItem("global-chat-name");
    if (name) form.reset({ name });
  }, []);

  const onSubmit = (data: z.infer<typeof formSchema>) => {
    localStorage.setItem("global-chat-name", data.name);
    onOpenChange(false);
  };

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Nama Pengguna</DialogTitle>
            <DialogDescription>
              Masukkan nama pengguna untuk mulai mengobrol di global chat.
            </DialogDescription>
          </DialogHeader>

          <Form {...form}>
            <form onSubmit={(e) => e.preventDefault()}>
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Nama</FormLabel>
                    <FormControl>
                      <Input placeholder="Masukkan nama" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </form>
          </Form>

          <DialogFooter>
            <DialogClose asChild>
              <Button>Batal</Button>
            </DialogClose>
            <Button onClick={form.handleSubmit(onSubmit)}>Simpan</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
};

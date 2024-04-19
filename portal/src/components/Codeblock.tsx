import { Box, Card, CardContent, Typography } from '@suid/material';
import * as Prism from 'prismjs';
import 'prismjs/components/prism-json';
import { Component } from "solid-js";

interface CodeblockProps {
    children: any;
    title: string;
    class?: string;
}

const Codeblock: Component<CodeblockProps> = (props: CodeblockProps) => {
    // Use Prism.js to highlight the code
    const html = Prism.highlight(props.children, Prism.languages.json, "json");

    return (
        <Card class={props.class ?? ""}>
            <CardContent>
                <Typography variant='body2' color='text.secondary'>
                    {props.title}
                </Typography>
                <Box class='mt-2' sx={{
                    backgroundColor: 'rgb(40, 44, 52)',
                    padding: '1rem',
                    borderRadius: '4px',
                    overflowX: 'auto',
                    border: '1px solid #0d0d0d',
                }}>
                    <pre>
                        <code class="language-json" innerHTML={html} />
                    </pre>
                </Box>
            </CardContent>
        </Card>
    );
};

export default Codeblock;
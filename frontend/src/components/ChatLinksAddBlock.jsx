import { useState } from 'react';
import { Flex, IconButton, Input, CellSimple, Avatar, Counter, Grid, Container } from "@maxhub/max-ui";
import { LuPlus, LuX, LuCheck, LuLink, LuMessageSquare, LuUsers, LuSettings, LuHouse } from "react-icons/lu";

const ChatLinksAddBlock = () => {
    const [isAdding, setIsAdding] = useState(false);
    const [title, setTitle] = useState('');
    const [subtitle, setSubtitle] = useState('');
    const [selectedIcon, setSelectedIcon] = useState('LuLink');

    const icons = [
        { name: 'LuLink', icon: LuLink },
        { name: 'LuMessageSquare', icon: LuMessageSquare },
        { name: 'LuUsers', icon: LuUsers },
        { name: 'LuSettings', icon: LuSettings },
        { name: 'LuHome', icon: LuHouse },
    ];

    const handleAdd = () => {
        console.log({ title, subtitle, icon: selectedIcon });

        setTitle('');
        setSubtitle('');
        setSelectedIcon('LuLink');
        setIsAdding(false);
    };

    const handleCancel = () => {
        setTitle('');
        setSubtitle('');
        setSelectedIcon('LuLink');
        setIsAdding(false);
    };

    return (
        <Container className='mt-1'>
            {isAdding ? (
                <Grid justify='center' cols={1} gap={12} style={{
                    padding: '16px',
                    background: 'var(--glass-bg)',
                    borderRadius: '12px',
                    border: '1px solid var(--glass-border)'
                }}>
                    <Flex justify='center' gap={12} wrap="wrap">
                        {icons.map(({ name, icon: Icon }) => (
                            <IconButton
                                key={name}
                                mode={selectedIcon === name ? 'primary' : 'secondary'}
                                appearance={selectedIcon === name ? 'themed' : 'neutral'}
                                onClick={() => setSelectedIcon(name)}
                            >
                                <Icon fontSize="1.5em" />
                            </IconButton>
                        ))}
                    </Flex>

                    <Input
                        placeholder="Введите заголовок"
                        mode='secondary'
                        value={title}
                        onChange={(e) => setTitle(e.target.value)}
                        size="large"
                    />

                    <Input
                        placeholder="Введите подзаголовок (необязательно)"
                        mode='secondary'
                        value={subtitle}
                        onChange={(e) => setSubtitle(e.target.value)}
                        size="large"
                    />

                    <Flex justify='center' gap={5}>
                        <IconButton
                            mode="secondary"
                            appearance="neutral"
                            onClick={handleCancel}
                        >
                            <LuX fontSize="1.5em" />
                        </IconButton>
                        <IconButton
                            mode="primary"
                            appearance="themed"
                            onClick={handleAdd}
                            disabled={!title.trim()}
                        >
                            <LuCheck fontSize="1.5em" />
                        </IconButton>
                    </Flex>
                </Grid>
            ) : (
                <Flex justify="center">
                    <IconButton
                        mode="secondary"
                        appearance="neutral"
                        onClick={() => setIsAdding(true)}
                    >
                        <LuPlus fontSize="2em" />
                    </IconButton>
                </Flex>
            )}
        </Container>
    );
};

export default ChatLinksAddBlock;

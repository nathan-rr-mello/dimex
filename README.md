# Trabalho 1 de Sistemas Distribuídos

# Alunos

- Gabriel Tabajara dos Santos
- Giovanni Masoni Schenato
- Nathan dos Reis Ramos de Mello
- João Pedro Fernandes Feijó

# Como rodar

- Em windows é possível rodar o trabalho apenas rodando o arquivo `run.bat`
  - Dessa forma também salvamos os logs dos processos em um arquivo `t<id>.txt`
  
- Caso de alguma falha no script ou o sistema operacional não for windows podemos rodar um processo por vez, utilizando o comando:
  
    `go run useDIMEX-f.go <id> 127.0.0.1:<porta>...`

    - Onde passamos como argumento o id do processo e as portas dele e dos outros processos.
    - O arquivo `useDimex-f.go` possui um exemplo dos comandos para 3 processos.
    - Antes de rodar dessa forma, delete todos os arquivos `snapshot<id>.txt`, pois os snapshot são adicionados a esses arquivos, então caso eles já tenham algum texto dentro esse texto não é deletado.

# Snapshots

- A cada 50ms o processo com id 0 inicia um snapshot, esses são gravados nos arquivos `snapshot<id>.txt`

# Teste com invariantes

- Ferramenta que avalia se para cada snapshot os estados estão consistentes, conforme as seguintes invariantes:
  
    1- No máximo um processo na SC.
    2- Se todos processos estão em "não quero a SC", então todos waitings tem que ser falsos e não deve haver mensagens.
    3- Se um processo q está marcado como waiting em p, então p está na SC ou quer a SC.

- Caso todos as invariantes sejam satisfeitas o algoritmo está funcionando corretamente, se uma delas falhar ele informará o primeiro snapshot id que não está consistente.

# Falhas

- No arquivo `Falhas.txt` temos um exemplo de duas mudanças que fariam o programa falhar, elas podem ser usadas para testas a ferramenta que avalia os snapshots.


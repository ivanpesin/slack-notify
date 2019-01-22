#!/usr/bin/perl
# vim: ts=2 sts=2 sw=2 et si nu

use Getopt::Std;
use File::Temp qw/ tempfile /;
use JSON;

=head1 NAME

slack2mail.pl - script forwards mail messages from local mailbox
                to specified slack team's channel

=head1 SYNOPSIS

  slack2mail.pl -c "#sysops" -w "https://hooks.slack.com/services/..."

=head1 DESCRIPTION

This scripts forwards mail messages from local mailbox to specified
slack team's channel. It uses C</bin/mail> to check, retrieve, and delete
messages from local mailbox; it uses C</usr/bin/curl> to trigger webhook
and send contents of the mail messages.

B<NOTE>: After the mail message is successfully sent to slack, it is
I<deleted> from the local mailbox.

=head2 Parameters

=over 10

=item C<-c>
Slack team's channel name.

=item C<-v>
Show debugging output.

=item C<-w>
Slack team's webhook URL.

=back

=cut

use strict;
use warnings;

my $slack_webhook = "https://hooks.slack.com/services/...";
my $slack_channel = "#random";

sub debug {
  if (defined($ENV{'DEBUG'}) && $ENV{'DEBUG'} eq "1") {
    my @args = @_;
    for (@args) {
      chomp;
      print "D: " . $_ . "\n";
    }
  }
}

sub run {
  my $cmd = shift @_;
  my @out = `$cmd 2>&1`;
  my $rc = $? >> 8;
  return ($rc, @out);
}

sub send2slack {
  #return (0, 200, ( "ok" ) );
  my $payload = shift @_;

  my ($fh, $fn) = tempfile();
  print $fh encode_json($payload);
  close($fh);

  debug(" => sending to slack ...");
  my ($rc, @out) = run("/usr/bin/curl -X POST -s -w '\\n\%{http_code}' --data-urlencode payload\@$fn $slack_webhook");
  my $code = -1;
  if ($out[$#out] =~ /(\d+)/) {
    $code = $1 ;
    splice @out, $#out - 1;
  }
  debug(" rc = $rc | code = $code | out = @out");
  return ($rc, $code, @out);
}

sub check_deps {
  unless (-x "/bin/mail") {
    print STDERR "E: /bin/mail is required\n";
    exit 1;
  }
  unless (-x "/usr/bin/curl") {
    print STDERR "E: /usr/bin/curl is required\n";
    exit 1;
  }
}

#
# --- main
#

my %opts;
getopts("f:vw:c:h",\%opts);

$ENV{'DEBUG'}=1 if defined($opts{v});
$slack_webhook = $opts{w} if defined($opts{w});
$slack_channel = $opts{c} if defined($opts{c});

check_deps();

my ($rc, @mail) = run("/bin/mail -H");
if ($rc == 0 && $#mail >= 0 && $mail[0] =~ /No mail for/) {
  debug "No mail";
  exit 0;
}
if ($rc != 0) {
  print STDERR "/bin/mail error: $rc\n@mail";
  exit 1;
}

unless (@mail) {
  debug "No mail";
  exit 0;
}

debug @mail;

for (@mail) {
  my ($rc, @msg) = run("/bin/echo type 1 | mail -N");
  if ($rc != 0) {
    print STDERR "/bin/mail error: $rc\n@msg";
    exit 1;
  }

  debug("Processing: $msg[0]");
  # remove leading and trailing lines added by mailx
  # pick up subject/to, split into 3k chunks (slack doesn't like bigger)
  #
  my ($subj, $to) = ( "", "" );
  my ($i,$sz,$part) = (0,3,"```");
  my @text;

  for (@msg) {
    $i++;
    next if ($i == 1);

    chomp;
    last if (/^Held \d+ messages? in /);

    if (/Subject: /) {
      $subj = $_;
      $subj =~ s/^Subject: //;
    }
        if (/To: /) {
      $to = $_;
      $to =~ s/^To: //;
    }

    $sz += length($_);
    $part .= "\n" . $_;
    if ($sz > 3000) {
      $part .= "```";
      $part = ":e-mail: *$to*: _${subj}_" . $part;
      push @text, $part;
      $part = "```";
      $sz = 3;
    }
  }
  $part .= "```";
  $part = ":e-mail: *$to*: _${subj}_" . $part;
  push @text, $part;

  debug(" -> $to | $subj");
  debug(" chunks: " . ($#text + 1));

  my $ok_to_rm = 1;
  for my $t (@text) {
    my %notification = (
      'channel' => $slack_channel,
      'username' => 'mail2slack',
      'text' => $t
    );

    my ($rc, $code, @out) = send2slack \%notification;
    if ($code != 200 || $rc != 0 ) {
      print STDERR "E: rc = $rc | code = $code\n";
      print STDERR "E: rcvd:\n@out\n";
      print STDERR "E: failed to send message part to slack, skipping ...";
      $ok_to_rm = 0;
      exit 1;
    }
  }

  if ($ok_to_rm) {
    debug("Message sent, will delete from mailbox");

    my ($rc, @msg) = run("/bin/echo d 1 | mail -N");
    if ($rc != 0) {
      print STDERR "/bin/mail error: $rc\n@msg";
      exit 1;
    }

    send2slack({ 'channel' => $slack_channel,'username' => 'mail2slack','text' => ":e-mail: :negative_squared_cross_mark: deleted - *$to*: _${subj}_" });
  }
}
